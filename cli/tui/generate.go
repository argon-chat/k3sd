package tui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rivo/tview"
)

var addonList = []string{
	"gitea", "gitea-ingress", "cert-manager", "traefik", "prometheus", "cluster-issuer", "linkerd", "linkerd-mc",
}

type AddonFormField struct {
	Label      string
	Default    string
	IsPassword bool
}

func addSectionToForm(form *tview.Form, fields []AddonFormField) {
	for _, f := range fields {
		if f.IsPassword {
			form.AddPasswordField(f.Label, f.Default, 20, '*', nil)
		} else {
			form.AddInputField(f.Label, f.Default, 20, nil, nil)
		}
	}
}

func addGiteaSectionToForm(app *tview.Application, enabledAddons map[string]bool, domain, masterIP, masterUser, masterPassword, nodeName string, privateNet bool, outputPath string, onDone func()) *tview.Form {
	giteaFields := []AddonFormField{
		{Label: "POSTGRES_USER", Default: "gitea"},
		{Label: "POSTGRES_PASSWORD", Default: "changeme", IsPassword: false},
		{Label: "POSTGRES_DB", Default: "giteadb"},
	}
	form := tview.NewForm()
	addSectionToForm(form, giteaFields)
	form.AddButton("Generate", func() {
		giteaSubs := map[string]string{
			"${POSTGRES_USER}":     form.GetFormItemByLabel("POSTGRES_USER").(*tview.InputField).GetText(),
			"${POSTGRES_PASSWORD}": form.GetFormItemByLabel("POSTGRES_PASSWORD").(*tview.InputField).GetText(),
			"${POSTGRES_DB}":       form.GetFormItemByLabel("POSTGRES_DB").(*tview.InputField).GetText(),
		}
		addons := buildAddonsMap(enabledAddons, domain, giteaSubs)
		cluster := buildClusterConfig(masterIP, masterUser, masterPassword, nodeName, domain, privateNet, addons)
		writeClusterConfigFile(app, []interface{}{cluster}, outputPath)
	})
	form.AddButton("Cancel", func() { app.Stop() })
	form.SetBorder(true).SetTitle("Gitea DB Configuration").SetTitleAlign(tview.AlignLeft)
	return form
}

func buildAddonCheckboxes(form *tview.Form) []*tview.Checkbox {
	checks := make([]*tview.Checkbox, len(addonList))
	for i, name := range addonList {
		cb := tview.NewCheckbox().SetLabel(name)
		checks[i] = cb
		form.AddFormItem(cb)
	}
	return checks
}

func buildAddonsMap(enabledAddons map[string]bool, domain string, giteaSubs map[string]string) map[string]interface{} {
	addons := make(map[string]interface{})
	for _, name := range addonList {
		if !enabledAddons[name] {
			continue
		}
		if name == "gitea" && giteaSubs != nil {
			addons[name] = map[string]interface{}{
				"enabled": true,
				"subs":    giteaSubs,
			}
		} else if name == "gitea-ingress" || name == "cluster-issuer" {
			addons[name] = map[string]interface{}{
				"enabled": true,
				"subs": map[string]string{
					"${DOMAIN}": domain,
				},
			}
		} else {
			addons[name] = map[string]interface{}{
				"enabled": true,
				"subs":    map[string]string{},
			}
		}
	}
	return addons
}

func buildClusterConfig(masterIP, masterUser, masterPassword, nodeName, domain string, privateNet bool, addons map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"address":    masterIP,
		"user":       masterUser,
		"password":   masterPassword,
		"nodeName":   nodeName,
		"domain":     domain,
		"privateNet": privateNet,
		"workers":    []interface{}{},
		"addons":     addons,
	}
}

func writeClusterConfigFile(app *tview.Application, clusters []interface{}, outputPath string) {
	f, err := os.Create(outputPath)
	if err != nil {
		modal := tview.NewModal().SetText(fmt.Sprintf("Failed to write file: %v", err)).AddButtons([]string{"Quit"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.Stop()
		})
		app.SetRoot(modal, true)
		return
	}
	// Close file and check error
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(clusters)
	if err != nil {
		modal := tview.NewModal().SetText(fmt.Sprintf("Failed to encode JSON: %v", err)).AddButtons([]string{"Quit"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.Stop()
		})
		app.SetRoot(modal, true)
		return
	}
	modal := tview.NewModal().SetText("Cluster config generated!\n" + outputPath).AddButtons([]string{"OK"}).SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		app.Stop()
	})
	app.SetRoot(modal, true)
}

func buildClusterForm(app *tview.Application) *tview.Form {
	var privateNet bool
	form := tview.NewForm()
	mainFields := []AddonFormField{
		{Label: "Master node IP", Default: ""},
		{Label: "Master SSH user", Default: ""},
		{Label: "Master SSH password", Default: "", IsPassword: true},
		{Label: "Master node name", Default: ""},
		{Label: "Cluster domain", Default: ""},
	}
	addSectionToForm(form, mainFields)
	form.AddCheckbox("Private network (privateNet)", false, func(checked bool) { privateNet = checked })
	addonChecks := buildAddonCheckboxes(form)
	outputPath := "clusters.generated.json"
	form.AddInputField("Output file path", outputPath, 40, nil, func(text string) {
		outputPath = text
	})
	form.AddButton("Next", func() {
		masterIP := form.GetFormItemByLabel("Master node IP").(*tview.InputField).GetText()
		masterUser := form.GetFormItemByLabel("Master SSH user").(*tview.InputField).GetText()
		masterPassword := form.GetFormItemByLabel("Master SSH password").(*tview.InputField).GetText()
		nodeName := form.GetFormItemByLabel("Master node name").(*tview.InputField).GetText()
		domain := form.GetFormItemByLabel("Cluster domain").(*tview.InputField).GetText()
		enabledAddons := map[string]bool{}
		for i, name := range addonList {
			if addonChecks[i].IsChecked() {
				enabledAddons[name] = true
			}
		}
		if enabledAddons["gitea"] {
			giteaForm := addGiteaSectionToForm(app, enabledAddons, domain, masterIP, masterUser, masterPassword, nodeName, privateNet, outputPath, func() { app.Stop() })
			app.SetRoot(giteaForm, true)
			return
		}
		addons := buildAddonsMap(enabledAddons, domain, nil)
		cluster := buildClusterConfig(masterIP, masterUser, masterPassword, nodeName, domain, privateNet, addons)
		writeClusterConfigFile(app, []interface{}{cluster}, outputPath)
	})
	form.AddButton("Cancel", func() { app.Stop() })
	form.SetBorder(true).SetTitle("k3sd Cluster Config Generator").SetTitleAlign(tview.AlignLeft)
	return form
}

func RunGenerateTUI() error {
	app := tview.NewApplication()
	form := buildClusterForm(app)
	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		return err
	}
	return nil
}
