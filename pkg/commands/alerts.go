package commands

import (
	"context"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/prometheus/alertmanager/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/grafana/cortextool/pkg/client"
	"github.com/grafana/cortextool/pkg/printer"
)

// AlertCommand configures and executes rule related cortex api operations
type AlertCommand struct {
	ClientConfig           client.Config
	AlertmanagerConfigFile string
	TemplateFiles          []string
	DisableColor           bool

	cli *client.CortexClient
}

// Register rule related commands and flags with the kingpin application
func (a *AlertCommand) Register(app *kingpin.Application) {
	alertCmd := app.Command("alertmanager", "View & edit alertmanager configs stored in cortex.").PreAction(a.setup)
	alertCmd.Flag("address", "Address of the cortex cluster, alternatively set CORTEX_ADDRESS.").Envar("CORTEX_ADDRESS").Required().StringVar(&a.ClientConfig.Address)
	alertCmd.Flag("id", "Cortex tenant id, alternatively set CORTEX_TENANT_ID.").Envar("CORTEX_TENANT_ID").Required().StringVar(&a.ClientConfig.ID)
	alertCmd.Flag("key", "Api key to use when contacting cortex, alternatively set CORTEX_API_KEY.").Default("").Envar("CORTEX_API_KEY").StringVar(&a.ClientConfig.Key)
	alertCmd.Flag("tls-ca-cert", "TLS CA certificate to verify cortex API as part of mTLS, alternatively set CORTEX_TLS_CA_CERT.").Default("").Envar("CORTEX_TLS_CA_CERT").StringVar(&a.ClientConfig.TLS.CAPath)
	alertCmd.Flag("tls-cert", "TLS client certificate to authenticate with cortex API as part of mTLS, alternatively set CORTEX_TLS_CLIENT_CERT.").Default("").Envar("CORTEX_TLS_CLIENT_CERT").StringVar(&a.ClientConfig.TLS.CertPath)
	alertCmd.Flag("tls-key", "TLS client certificate private key to authenticate with cortex API as part of mTLS, alternatively set CORTEX_TLS_CLIENT_KEY.").Default("").Envar("CORTEX_TLS_CLIENT_KEY").StringVar(&a.ClientConfig.TLS.KeyPath)

	// Get Alertmanager Configs Command
	getAlertsCmd := alertCmd.Command("get", "Get the alertmanager config currently in the cortex alertmanager.").Action(a.getConfig)
	getAlertsCmd.Flag("disable-color", "disable colored output").BoolVar(&a.DisableColor)

	alertCmd.Command("delete", "Delete the alertmanager config currently in the cortex alertmanager.").Action(a.deleteConfig)

	loadalertCmd := alertCmd.Command("load", "load a set of rules to a designated cortex endpoint").Action(a.loadConfig)
	loadalertCmd.Arg("config", "alertmanager configuration to load").Required().StringVar(&a.AlertmanagerConfigFile)
	loadalertCmd.Arg("template-files", "The template files to load").ExistingFilesVar(&a.TemplateFiles)
}

func (a *AlertCommand) setup(k *kingpin.ParseContext) error {
	cli, err := client.New(a.ClientConfig)
	if err != nil {
		return err
	}
	a.cli = cli

	return nil
}

func (a *AlertCommand) getConfig(k *kingpin.ParseContext) error {
	cfg, templates, err := a.cli.GetAlertmanagerConfig(context.Background())
	if err != nil {
		if err == client.ErrResourceNotFound {
			log.Infof("no alertmanager config currently exist for this user")
			return nil
		}
		return err
	}

	p := printer.New(a.DisableColor)

	return p.PrintAlertmanagerConfig(cfg, templates)
}

func (a *AlertCommand) loadConfig(k *kingpin.ParseContext) error {
	content, err := ioutil.ReadFile(a.AlertmanagerConfigFile)
	if err != nil {
		return errors.Wrap(err, "unable to load config file: "+a.AlertmanagerConfigFile)
	}

	cfg := string(content)
	_, err = config.Load(cfg)
	if err != nil {
		return err
	}

	templates := map[string]string{}
	for _, f := range a.TemplateFiles {
		tmpl, err := ioutil.ReadFile(f)
		if err != nil {
			return errors.Wrap(err, "unable to load template file: "+f)
		}
		templates[f] = string(tmpl)
	}

	return a.cli.CreateAlertmanagerConfig(context.Background(), cfg, templates)
}

func (a *AlertCommand) deleteConfig(k *kingpin.ParseContext) error {
	err := a.cli.DeleteAlermanagerConfig(context.Background())
	if err != nil && err != client.ErrResourceNotFound {
		return err
	}
	return nil
}
