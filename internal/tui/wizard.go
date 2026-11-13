package tui

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/hip/internal/cache"
	"github.com/mipmip/hip/internal/config"
	"github.com/mipmip/hip/internal/creator"
	"github.com/mipmip/hip/internal/provider"
	"github.com/mipmip/hip/internal/syncer"
)

type wizardStep int

const (
	wizardProvider wizardStep = iota
	wizardOwner
	wizardName
	wizardVisibility
	wizardReadme
	wizardLicense
	wizardCheckout
	wizardHolder
	wizardReview
)

//go:embed licenses/apache-2.0.txt
var apacheLicense string

//go:embed licenses/gpl-3.0.txt
var gplLicense string

const mitLicense = `MIT License

Copyright (c) %d %s

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

var licenseChoices = []string{"none", "MIT", "Apache-2.0", "GPL-3.0"}

type wizard struct {
	step wizardStep

	provider       *config.Provider
	providerCursor int
	owner          string
	ownerCursor    int
	owners         []string

	name      textinput.Model
	desc      textinput.Model
	nameFocus bool

	visibility int
	readme     bool
	readmePath string
	readmeText string
	license    int
	checkout   bool
	holder     textinput.Model

	busy bool
	err  string
}

func newWizard() *wizard {
	w := &wizard{
		name:      textinput.New(),
		desc:      textinput.New(),
		holder:    textinput.New(),
		nameFocus: true,
	}
	w.name.Placeholder = "repository name"
	w.name.Prompt = "name: "
	w.desc.Placeholder = "description (optional)"
	w.desc.Prompt = "description: "
	w.holder.Placeholder = "copyright holder"
	w.holder.Prompt = "holder: "
	w.name.Focus()
	return w
}

func (m *Model) newRepo() (tea.Model, tea.Cmd) {
	switch m.nav {
	case levelProviders, levelOwners, levelRepos:
	default:
		return m, nil
	}
	w := newWizard()
	m.wizard = w

	if m.nav == levelRepos {
		if it, ok := m.currentRepo(); ok {
			w.provider = it.Provider
			w.owner = it.Repo.Owner
			w.step = wizardName
			w.name.Focus()
			return m, nil
		}
	}
	if m.selProvider != nil {
		w.provider = m.selProvider
		w.owners = m.ownersFor(m.selProvider)
		if len(w.owners) == 0 && m.selProvider.Username != "" {
			w.owners = []string{m.selProvider.Username}
		}
		if m.selOwner != "" {
			w.owner = m.selOwner
			for i, owner := range w.owners {
				if owner == m.selOwner {
					w.ownerCursor = i
					break
				}
			}
			w.step = wizardName
			w.name.Focus()
			return m, nil
		}
		w.step = wizardOwner
		return m, nil
	}
	w.step = wizardProvider
	return m, nil
}

func (m *Model) handleWizardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	w := m.wizard
	if w == nil {
		return m, nil
	}
	if w.busy {
		return m, nil
	}
	if msg.Type == tea.KeyCtrlC {
		m.quitting = true
		return m, tea.Quit
	}
	if msg.Type == tea.KeyEsc {
		m.wizard = nil
		m.status = "create canceled"
		return m, nil
	}

	switch w.step {
	case wizardProvider:
		switch msg.String() {
		case "up", "k":
			if w.providerCursor > 0 {
				w.providerCursor--
			}
		case "down", "j":
			if w.providerCursor < len(m.providers)-1 {
				w.providerCursor++
			}
		case "enter":
			if len(m.providers) == 0 {
				w.err = "no providers configured"
				return m, nil
			}
			w.provider = m.providers[w.providerCursor]
			w.owners = m.ownersFor(w.provider)
			if len(w.owners) == 0 && w.provider.Username != "" {
				w.owners = []string{w.provider.Username}
			}
			w.ownerCursor = 0
			w.owner = ""
			w.step = wizardOwner
		}
		return m, nil

	case wizardOwner:
		switch msg.String() {
		case "up", "k":
			if w.ownerCursor > 0 {
				w.ownerCursor--
			}
		case "down", "j":
			if w.ownerCursor < len(w.owners)-1 {
				w.ownerCursor++
			}
		case "enter":
			if len(w.owners) == 0 {
				w.err = "no owners available"
				return m, nil
			}
			w.owner = w.owners[w.ownerCursor]
			w.step = wizardName
			w.name.Focus()
		}
		return m, nil

	case wizardName:
		if msg.Type == tea.KeyTab || msg.Type == tea.KeyShiftTab {
			w.nameFocus = !w.nameFocus
			if w.nameFocus {
				w.desc.Blur()
				w.name.Focus()
			} else {
				w.name.Blur()
				w.desc.Focus()
			}
			return m, nil
		}
		if msg.Type == tea.KeyEnter {
			name := strings.TrimSpace(w.name.Value())
			if name == "" {
				w.err = "name is required"
				return m, nil
			}
			if strings.ContainsAny(name, "/ \t") {
				w.err = "name must not contain spaces or slashes"
				return m, nil
			}
			w.name.SetValue(name)
			w.err = ""
			w.step = wizardVisibility
			return m, nil
		}
		var cmd tea.Cmd
		if w.nameFocus {
			w.name, cmd = w.name.Update(msg)
		} else {
			w.desc, cmd = w.desc.Update(msg)
		}
		return m, cmd

	case wizardVisibility:
		switch msg.String() {
		case "left", "h", "right", "l", " ":
			w.visibility = 1 - w.visibility
		case "enter":
			w.step = wizardReadme
		}
		return m, nil

	case wizardReadme:
		switch msg.String() {
		case "left", "h", "right", "l", " ":
			w.readme = !w.readme
		case "enter":
			if !w.readme {
				w.step = wizardLicense
				return m, nil
			}
			return m.startREADME()
		}
		return m, nil

	case wizardLicense:
		switch msg.String() {
		case "up", "k":
			if w.license > 0 {
				w.license--
			}
		case "down", "j":
			if w.license < len(licenseChoices)-1 {
				w.license++
			}
		case "enter":
			w.step = wizardCheckout
		}
		return m, nil

	case wizardCheckout:
		switch msg.String() {
		case "left", "h", "right", "l", " ":
			w.checkout = !w.checkout
		case "enter":
			if licenseChoices[w.license] != "none" && strings.TrimSpace(w.holder.Value()) == "" {
				w.holder.SetValue(gitUserName())
				if strings.TrimSpace(w.holder.Value()) == "" {
					w.step = wizardHolder
					w.holder.Focus()
					return m, nil
				}
			}
			w.step = wizardReview
		}
		return m, nil

	case wizardHolder:
		if msg.Type == tea.KeyEnter {
			if strings.TrimSpace(w.holder.Value()) == "" {
				w.err = "holder is required for a license"
				return m, nil
			}
			w.holder.SetValue(strings.TrimSpace(w.holder.Value()))
			w.err = ""
			w.step = wizardReview
			return m, nil
		}
		var cmd tea.Cmd
		w.holder, cmd = w.holder.Update(msg)
		return m, cmd

	case wizardReview:
		if msg.String() == "enter" {
			if w.provider == nil || w.owner == "" || strings.TrimSpace(w.name.Value()) == "" {
				w.err = "incomplete repository details"
				return m, nil
			}
			w.busy = true
			w.err = ""
			return m, m.wizardCreateCmd()
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) startREADME() (tea.Model, tea.Cmd) {
	w := m.wizard
	name := strings.TrimSpace(w.name.Value())
	content := "# " + name + "\n"
	if desc := strings.TrimSpace(w.desc.Value()); desc != "" {
		content += "\n" + desc + "\n"
	}
	f, err := os.CreateTemp("", "hip-readme-*.md")
	if err != nil {
		w.err = "creating README: " + err.Error()
		return m, nil
	}
	path := f.Name()
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(path)
		w.err = "writing README: " + err.Error()
		return m, nil
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		w.err = "writing README: " + err.Error()
		return m, nil
	}
	w.readmePath = path
	return m, editREADME(path)
}

func editREADME(path string) tea.Cmd {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		editor = "vi"
	}
	fields := strings.Fields(editor)
	args := append(fields[1:], path)
	cmd := exec.Command(fields[0], args...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return wizardEditorDoneMsg{Path: path, Err: err}
	})
}

type wizardEditorDoneMsg struct {
	Path string
	Err  error
}

func (m *Model) finishREADME(msg wizardEditorDoneMsg) (tea.Model, tea.Cmd) {
	w := m.wizard
	if w == nil {
		os.Remove(msg.Path)
		return m, nil
	}
	defer os.Remove(msg.Path)
	if msg.Err != nil {
		m.wizard = nil
		m.status = "create canceled: editor exited with an error"
		return m, nil
	}
	content, err := os.ReadFile(msg.Path)
	if err != nil {
		w.err = "reading README: " + err.Error()
		return m, nil
	}
	w.readmePath = ""
	w.readmeText = string(content)
	w.step = wizardLicense
	return m, nil
}

func gitUserName() string {
	out, err := exec.Command("git", "config", "--get", "user.name").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func licenseText(choice, holder string) string {
	year := time.Now().Year()
	switch choice {
	case "MIT":
		return fmt.Sprintf(mitLicense, year, holder)
	case "Apache-2.0":
		return fmt.Sprintf("SPDX-License-Identifier: Apache-2.0\nCopyright (c) %d %s\n\n%s", year, holder, apacheLicense)
	case "GPL-3.0":
		return fmt.Sprintf("SPDX-License-Identifier: GPL-3.0-only\nCopyright (c) %d %s\n\n%s", year, holder, gplLicense)
	default:
		return ""
	}
}

type wizardCreatedMsg struct {
	Result   creator.Result
	Err      error
	Checkout bool
}

func (m *Model) wizardCreateCmd() tea.Cmd {
	w := m.wizard
	providerCfg := w.provider
	opts := provider.CreateRepoOptions{
		Owner:       w.owner,
		Name:        strings.TrimSpace(w.name.Value()),
		Description: strings.TrimSpace(w.desc.Value()),
		Visibility:  "private",
	}
	if w.visibility == 1 {
		opts.Visibility = "public"
	}
	req := creator.Request{
		ProviderName: providerCfg.Name,
		Options:      opts,
		README:       w.readmeText,
		License:      licenseText(licenseChoices[w.license], w.holder.Value()),
	}
	checkout := w.checkout
	run := m.wizardRun
	return func() tea.Msg {
		res, err := run(providerCfg, req)
		return wizardCreatedMsg{Result: res, Err: err, Checkout: checkout}
	}
}

func defaultWizardRun(p *config.Provider, req creator.Request) (creator.Result, error) {
	remote, err := provider.NewDefaultRegistry().Build(p)
	if err != nil {
		return creator.Result{}, err
	}
	eng := creator.NewEngine(remote, cacheStore{}, func(r provider.Repo) string {
		return syncer.CloneURL(p, r)
	})
	return eng.Create(context.Background(), req)
}

type cacheStore struct{}

func (cacheStore) AddRepo(providerName string, repo provider.Repo) error {
	return cache.AddRepo(providerName, repo)
}

func (m *Model) handleWizardCreated(msg wizardCreatedMsg) (tea.Model, tea.Cmd) {
	w := m.wizard
	if w == nil {
		return m, nil
	}
	if msg.Err != nil {
		w.busy = false
		w.err = msg.Err.Error()
		return m, nil
	}
	p := w.provider
	checkout := msg.Checkout
	m.wizard = nil
	m.loadProviderInto(p)
	m.selectCreatedRepo(p, msg.Result.Repo)
	m.status = "created " + msg.Result.Repo.Owner + "/" + msg.Result.Repo.Name
	if checkout {
		return m.clone()
	}
	return m, nil
}

func (m *Model) selectCreatedRepo(p *config.Provider, repo provider.Repo) {
	m.nav = levelRepos
	m.flatAll = false
	m.selProvider = p
	m.selOwner = repo.Owner
	m.selColl = ""
	m.filtering = false
	m.filter.SetValue("")
	m.cursor = 0
	m.offset = 0
	for i, it := range m.visibleRepos() {
		if strings.EqualFold(it.Repo.Owner, repo.Owner) && strings.EqualFold(it.Repo.Name, repo.Name) {
			m.cursor = i
			break
		}
	}
	m.clampCursor()
}

func (m *Model) wizardView() string {
	w := m.wizard
	if w == nil {
		return ""
	}
	if w.busy {
		return centerBox(m, "New repository\n\nCreating "+w.owner+"/"+w.name.Value()+"...")
	}
	var b strings.Builder
	b.WriteString("New repository\n\n")
	switch w.step {
	case wizardProvider:
		b.WriteString("Provider\n")
		for i, p := range m.providers {
			if i == w.providerCursor {
				b.WriteString("> " + p.Name + "\n")
			} else {
				b.WriteString("  " + p.Name + "\n")
			}
		}
		b.WriteString("\nup/down: choose  enter: next  esc: cancel")
	case wizardOwner:
		b.WriteString("Owner\n")
		for i, owner := range w.owners {
			if i == w.ownerCursor {
				b.WriteString("> " + owner + "\n")
			} else {
				b.WriteString("  " + owner + "\n")
			}
		}
		b.WriteString("\nup/down: choose  enter: next  esc: cancel")
	case wizardName:
		b.WriteString(w.name.View() + "\n" + w.desc.View() + "\n")
		b.WriteString("\ntab: switch field  enter: next  esc: cancel")
	case wizardVisibility:
		b.WriteString("Visibility\n")
		if w.visibility == 0 {
			b.WriteString("> private\n  public\n")
		} else {
			b.WriteString("  private\n> public\n")
		}
		b.WriteString("\nleft/right: choose  enter: next  esc: cancel")
	case wizardReadme:
		b.WriteString("README\n")
		if w.readme {
			b.WriteString("> yes, edit in $EDITOR\n  no\n")
		} else {
			b.WriteString("> no\n  yes, edit in $EDITOR\n")
		}
		b.WriteString("\nleft/right: choose  enter: next  esc: cancel")
	case wizardLicense:
		b.WriteString("License\n")
		for i, choice := range licenseChoices {
			if i == w.license {
				b.WriteString("> " + choice + "\n")
			} else {
				b.WriteString("  " + choice + "\n")
			}
		}
		b.WriteString("\nup/down: choose  enter: next  esc: cancel")
	case wizardCheckout:
		b.WriteString("Checkout\n")
		if w.checkout {
			b.WriteString("> yes, clone after creating\n  no\n")
		} else {
			b.WriteString("> no\n  yes, clone after creating\n")
		}
		b.WriteString("\nleft/right: choose  enter: next  esc: cancel")
	case wizardHolder:
		b.WriteString("Copyright holder\n" + w.holder.View() + "\n")
		b.WriteString("\nenter: next  esc: cancel")
	case wizardReview:
		visibility := "private"
		if w.visibility == 1 {
			visibility = "public"
		}
		readme := "no"
		if w.readme {
			readme = "yes"
		}
		checkout := "no"
		if w.checkout {
			checkout = "yes"
		}
		b.WriteString("Review\n\n")
		b.WriteString("provider:    " + w.provider.Name + "\n")
		b.WriteString("owner:       " + w.owner + "\n")
		b.WriteString("name:        " + w.name.Value() + "\n")
		if desc := strings.TrimSpace(w.desc.Value()); desc != "" {
			b.WriteString("description: " + desc + "\n")
		}
		b.WriteString("visibility:  " + visibility + "\n")
		b.WriteString("README:      " + readme + "\n")
		b.WriteString("license:     " + licenseChoices[w.license] + "\n")
		b.WriteString("checkout:    " + checkout + "\n")
		b.WriteString("\nenter: create  esc: cancel")
	}
	if w.err != "" {
		b.WriteString("\n\nerror: " + w.err)
	}
	return centerBox(m, b.String())
}
