package tui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/argon-chat/k3sd/pkg/types"
	"github.com/rivo/tview"
)

// --- Field Definitions ---

var clusterFields = []FieldDef{
	{"Master node IP", "", false},
	{"Master SSH user", "", false},
	{"Master SSH password", "", true},
	{"Master node name", "", false},
	{"Cluster domain", "", false},
	{"context", "k3sd", false},
}

var addonList = []string{
	"gitea", "gitea-ingress", "cert-manager", "traefik", "prometheus", "cluster-issuer", "linkerd", "linkerd-mc",
}

var giteaFields = []FieldDef{
	{"POSTGRES_USER", "gitea", false},
	{"POSTGRES_PASSWORD", "changeme", true},
	{"POSTGRES_DB", "giteadb", false},
}

// --- Types ---

type FieldDef struct {
	Label      string
	Default    string
	IsPassword bool
}

// --- Helpers ---

func showErrorModal(app *tview.Application, msg string) {
	modal := tview.NewModal().SetText(msg).AddButtons([]string{"OK"}).SetDoneFunc(func(_ int, _ string) { app.Stop() })
	app.SetRoot(modal, true)
}

func getInput(form *tview.Form, label string) string {
	field := form.GetFormItemByLabel(label)
	if field == nil {
		return ""
	}
	return field.(*tview.InputField).GetText()
}

func addFields(form *tview.Form, fields []FieldDef) {
	for _, f := range fields {
		if f.IsPassword {
			form.AddPasswordField(f.Label, f.Default, 20, '*', nil)
		} else {
			form.AddInputField(f.Label, f.Default, 20, nil, nil)
		}
	}
}

// --- Generic Form Builders ---

func buildAddonSubsForm(app *tview.Application, title string, fields []FieldDef, onBack func(), onDone func(subs map[string]string)) *tview.Form {
	form := tview.NewForm()
	addFields(form, fields)
	form.AddButton("Back", func() { onBack() })
	form.AddButton("Done", func() {
		subs := make(map[string]string)
		for _, f := range fields {
			subs[f.Label] = getInput(form, f.Label)
		}
		onDone(subs)
	})
	form.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignLeft)
	return form
}

// --- Addon Forms ---

func buildGiteaForm(app *tview.Application, onBack func(), onDone func(subs map[string]string)) *tview.Form {
	return buildAddonSubsForm(app, "Gitea Addon Configuration", giteaFields, onBack, onDone)
}

// --- Main Cluster Form ---

func buildClusterForm(app *tview.Application, onDone func(cluster *types.Cluster, outputPath string)) *tview.Form {
	form := tview.NewForm()
	addFields(form, clusterFields)
	var privateNet bool
	form.AddCheckbox("Private network (privateNet)", false, func(checked bool) { privateNet = checked })
	addonChecks := make([]*tview.Checkbox, len(addonList))
	form.AddTextView("", "Select addons:", 20, 1, false, false)
	for i, name := range addonList {
		cb := tview.NewCheckbox().SetLabel(name)
		addonChecks[i] = cb
		form.AddFormItem(cb)
	}
	outputPath := "clusters.generated.json"
	form.AddInputField("Output file path", outputPath, 40, nil, func(text string) { outputPath = text })
	form.AddButton("Next", func() {
		values := make(map[string]string)
		for _, f := range clusterFields {
			v := getInput(form, f.Label)
			if v == "" {
				showErrorModal(app, "Missing required field: "+f.Label)
				return
			}
			values[f.Label] = v
		}
		cluster := &types.Cluster{
			Worker: types.Worker{
				Address:  values["Master node IP"],
				User:     values["Master SSH user"],
				Password: values["Master SSH password"],
				NodeName: values["Master node name"],
			},
			Domain:     values["Cluster domain"],
			Context:    values["context"],
			PrivateNet: privateNet,
			Workers:    []types.Worker{},
			Addons:     map[string]types.AddonConfig{},
		}
		enabledAddons := map[string]bool{}
		for i, name := range addonList {
			if addonChecks[i].IsChecked() {
				enabledAddons[name] = true
			}
		}
		if enabledAddons["gitea"] {
			app.SetRoot(buildGiteaForm(app, func() {
				app.SetRoot(buildClusterForm(app, onDone), true)
			}, func(giteaSubs map[string]string) {
				for name := range enabledAddons {
					if name == "gitea" {
						cluster.Addons["gitea"] = types.AddonConfig{
							Enabled: true,
							Subs:    giteaSubs,
						}
					} else {
						cluster.Addons[name] = types.AddonConfig{Enabled: true}
					}
				}
				onDone(cluster, outputPath)
			}), true)
			return
		}
		for name := range enabledAddons {
			cluster.Addons[name] = types.AddonConfig{Enabled: true}
		}
		onDone(cluster, outputPath)
	})
	form.AddButton("Cancel", func() { app.Stop() })
	form.SetBorder(true).SetTitle("k3sd Cluster Config Generator").SetTitleAlign(tview.AlignLeft)
	return form
}

// --- Output ---

func writeClustersToFile(app *tview.Application, clusters []*types.Cluster, outputPath string) {
	f, err := os.Create(outputPath)
	if err != nil {
		showErrorModal(app, fmt.Sprintf("Failed to write file: %v", err))
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(clusters); err != nil {
		showErrorModal(app, fmt.Sprintf("Failed to encode JSON: %v", err))
		return
	}
	modal := tview.NewModal().SetText("Cluster config generated!\n" + outputPath).AddButtons([]string{"OK"}).SetDoneFunc(func(_ int, _ string) { app.Stop() })
	app.SetRoot(modal, true)
}

// --- Entry Point ---

func RunGenerateTUI() error {
	app := tview.NewApplication()
	form := buildClusterForm(app, func(cluster *types.Cluster, outputPath string) {
		writeClustersToFile(app, []*types.Cluster{cluster}, outputPath)
	})
	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		return err
	}
	return nil
}
