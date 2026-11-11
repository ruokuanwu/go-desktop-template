package view

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/lobehub/lobe-mygo/pkg/action"
	"github.com/lobehub/lobe-mygo/pkg/choice"
	"github.com/lobehub/lobe-mygo/pkg/field"
	"github.com/lobehub/lobe-mygo/pkg/theme"

	"go-desktop-template/internal/model"
	"go-desktop-template/internal/service"
)

type Preferences struct {
	Appearance string `json:"appearance"`
}

func LoadPreferences(dir string) (Preferences, error) {
	p := Preferences{Appearance: "system"}
	data, err := os.ReadFile(filepath.Join(dir, "preferences.json"))
	if os.IsNotExist(err) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return p, err
	}
	switch p.Appearance {
	case "system", "light", "dark":
	default:
		return p, fmt.Errorf("invalid appearance: %s", p.Appearance)
	}
	return p, nil
}

type View struct {
	todos       *service.Todos
	items       []model.Todo
	preferences Preferences
	dataDir     string
	page        string
	filter      string
	search      string
	title       string
	message     string
	busy        bool
	window      *mygo.Window
	ctx         context.Context
	cancel      context.CancelFunc
	workers     sync.WaitGroup
}

func New(todos *service.Todos, items []model.Todo, preferences Preferences, dataDir string) *View {
	ctx, cancel := context.WithCancel(context.Background())
	return &View{todos: todos, items: items, preferences: preferences, dataDir: dataDir, page: "任务", filter: "全部", ctx: ctx, cancel: cancel}
}

func (v *View) Attach(window *mygo.Window) { v.window = window }
func (v *View) Cancel()                    { v.cancel() }
func (v *View) Wait()                      { v.workers.Wait() }

// All view state is owned by the UI thread; workers publish only through Update.
func (v *View) run(job func(context.Context) error, success func()) {
	if v.busy {
		return
	}
	v.busy, v.message = true, ""
	v.workers.Add(1)
	go func() {
		defer v.workers.Done()
		err := job(v.ctx)
		var items []model.Todo
		if err == nil {
			items, err = v.todos.List(v.ctx)
		}
		v.window.Update(func() {
			if v.ctx.Err() != nil {
				return
			}
			v.busy = false
			if err != nil {
				v.message = err.Error()
				return
			}
			v.items = items
			if success != nil {
				success()
			}
		})
	}()
}

func (v *View) add() {
	title := v.title
	v.run(func(ctx context.Context) error { _, err := v.todos.Create(ctx, title); return err }, func() { v.title = "" })
}

func (v *View) setAppearance(appearance string) {
	p := Preferences{Appearance: appearance}
	dir := v.dataDir
	v.run(func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		data, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}
		temp := filepath.Join(dir, "preferences.json.tmp")
		if err := os.WriteFile(temp, data, 0o600); err != nil {
			return err
		}
		return os.Rename(temp, filepath.Join(dir, "preferences.json"))
	}, func() { v.preferences = p })
}

func (v *View) Render(c *ui.Context) {
	switch v.preferences.Appearance {
	case "light":
		c.SetTheme(theme.Default().Theme(false))
	case "dark":
		c.SetTheme(theme.Default().Theme(true))
	default:
		theme.Apply(c)
	}
	tk := theme.TokensOf(c)
	ui.Row(c).Fill().Background(tk.BgLayout).Children(func() {
		ui.Column(c).Width(176).FillHeight().Padding(20).Gap(12).Background(tk.BgContainer).Children(func() {
			ui.Text(c, "TODO DESKTOP").FontSize(16).Bold()
			ui.Text(c, "本地 · 原生 · 离线").FontSize(12).TextColor(tk.TextSecondary)
			for _, page := range []string{"概览", "任务", "设置"} {
				variant := action.Text
				if v.page == page {
					variant = action.Filled
				}
				c.Key(page)
				action.Button(c, page, action.ButtonProps{Variant: variant}).FillWidth().OnClick(func() { v.page = page })
			}
			ui.Column(c).Grow(1)
			ui.Text(c, "MyGo + Lobe MyGo").FontSize(11).TextColor(tk.TextSecondary)
		})
		ui.Column(c).Grow(1).MinWidth(0).FillHeight().Padding(28).Gap(18).Children(func() {
			ui.Text(c, v.page).FontSize(28).Bold()
			if v.message != "" {
				ui.Text(c, v.message).TextColor(tk.Error).Wrap()
			}
			if v.busy {
				ui.Text(c, "正在保存…").TextColor(tk.TextSecondary)
			}
			switch v.page {
			case "概览":
				v.dashboard(c)
			case "设置":
				v.settings(c)
			default:
				v.tasks(c)
			}
		})
	})
}

func (v *View) counts() (total, completed int) {
	for _, todo := range v.items {
		if todo.Completed {
			completed++
		}
	}
	return len(v.items), completed
}

func (v *View) dashboard(c *ui.Context) {
	total, completed := v.counts()
	tk := theme.TokensOf(c)
	ui.Text(c, "每一件小事，都值得完成。").TextColor(tk.TextSecondary)
	for _, stat := range []struct {
		label string
		count int
	}{{"全部任务", total}, {"待完成", total - completed}, {"已完成", completed}} {
		ui.Row(c).FillWidth().Padding(18).Radius(12).Background(tk.BgContainer).Gap(20).Children(func() {
			ui.Text(c, stat.label).Grow(1)
			ui.Text(c, fmt.Sprint(stat.count)).FontSize(26).Bold()
		})
	}
	action.Button(c, "管理任务", action.ButtonProps{Variant: action.Primary}).OnClick(func() { v.page = "任务" })
}

func (v *View) tasks(c *ui.Context) {
	total, completed := v.counts()
	tk := theme.TokensOf(c)
	ui.Text(c, fmt.Sprintf("共 %d 项 · 待完成 %d 项 · 已完成 %d 项", total, total-completed, completed)).TextColor(tk.TextSecondary)
	ui.Row(c).FillWidth().Gap(12).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			c.Key("new-todo")
			field.Input(c, &v.title, field.InputProps{Placeholder: "添加任务，按 Enter 保存（1–200 字）"}).Label("任务标题").Disabled(v.busy).OnSubmit(v.add)
		})
		action.Button(c, "添加", action.ButtonProps{Variant: action.Primary}).Disabled(v.busy).OnClick(v.add)
	})
	c.Key("search")
	field.Input(c, &v.search, field.InputProps{Placeholder: "搜索任务标题"})
	ui.Row(c).Gap(8).Children(func() {
		for _, filter := range []string{"全部", "待完成", "已完成"} {
			variant := action.Text
			if v.filter == filter {
				variant = action.Filled
			}
			c.Key(filter)
			action.Button(c, filter, action.ButtonProps{Variant: variant}).OnClick(func() { v.filter = filter })
		}
	})
	ui.Scroll(c).Grow(1).FillWidth().Children(func() {
		ui.Column(c).FillWidth().Gap(10).Children(func() {
			visible := 0
			for _, todo := range v.items {
				if v.filter == "待完成" && todo.Completed || v.filter == "已完成" && !todo.Completed {
					continue
				}
				if !strings.Contains(strings.ToLower(todo.Title), strings.ToLower(strings.TrimSpace(v.search))) {
					continue
				}
				visible++
				c.Key(todo.ID)
				ui.Row(c).FillWidth().Padding(14).Gap(12).Radius(10).Background(tk.BgContainer).Children(func() {
					checked := todo.Completed
					c.Key("completed")
					choice.Checkbox(c, &checked, "", choice.CheckboxProps{}).Label("切换完成状态").Disabled(v.busy).OnClick(func() {
						v.run(func(ctx context.Context) error { return v.todos.SetCompleted(ctx, todo.ID, !todo.Completed) }, nil)
					})
					color := tk.Text
					if todo.Completed {
						color = tk.TextSecondary
					}
					ui.Text(c, todo.Title).Grow(1).MinWidth(0).Wrap().TextColor(color)
					c.Key("delete")
					action.Button(c, "删除", action.ButtonProps{Variant: action.Text}).Disabled(v.busy).OnClick(func() {
						v.run(func(ctx context.Context) error { return v.todos.Delete(ctx, todo.ID) }, nil)
					})
				})
			}
			if visible == 0 {
				ui.Text(c, "暂无任务，添加一件想完成的事吧。").Padding(24).TextColor(tk.TextSecondary)
			}
		})
	})
}

func (v *View) settings(c *ui.Context) {
	tk := theme.TokensOf(c)
	ui.Text(c, "外观").FontSize(18).Bold()
	ui.Text(c, "主题设置自动保存在本机。").TextColor(tk.TextSecondary)
	ui.Row(c).Gap(10).Children(func() {
		for _, option := range []struct{ label, value string }{{"跟随系统", "system"}, {"浅色", "light"}, {"深色", "dark"}} {
			variant := action.Default
			if v.preferences.Appearance == option.value {
				variant = action.Primary
			}
			c.Key(option.value)
			action.Button(c, option.label, action.ButtonProps{Variant: variant}).Disabled(v.busy).OnClick(func() { v.setAppearance(option.value) })
		}
	})
	ui.Text(c, "本地数据目录").FontSize(18).Bold()
	ui.Text(c, v.dataDir).Wrap().TextColor(tk.TextSecondary)
	ui.Text(c, "任务保存在 todos.db；关闭程序后可复制此文件备份。").Wrap().TextColor(tk.TextSecondary)
	ui.Text(c, "纯 Go 原生界面，无 WebView，无本地 HTTP 服务。").Wrap().TextColor(tk.TextSecondary)
}
