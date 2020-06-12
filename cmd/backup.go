package cmd

import (
	"encoding/json"
	"flag"
	"github.com/pkg/errors"
	"github.com/rancher/norman/types/convert"
	"github.com/urfave/cli"
	"io/ioutil"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	//"os"
)

func BackupCommand() cli.Command {
	backupFlags := []cli.Flag{
		formatFlag,
		cli.StringFlag{
			Name:  "kubeconfig",
			Usage: "Pass kubeconfig of cluster to be backed up",
		},
	}

	return cli.Command{
		Name:   "backup",
		Usage:  "Operations with backups",
		Action: defaultAction(backupCreate),
		Flags:  backupFlags,
		Subcommands: []cli.Command{
			cli.Command{
				Name:        "create",
				Usage:       "Perform backup/create snapshot",
				Description: "\nCreate a backup of Rancher MCM",
				ArgsUsage:   "None",
				Action:      backupCreate,
				Flags:       backupFlags,
			},
		},
	}
}

func backupCreate(ctx *cli.Context) error {
	kubeconfig := ctx.String("kubeconfig")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}
	clientSet, err := clientset.NewForConfig(config)
	if err != nil {
		return err
	}
	CRDs, err := clientSet.ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
	if err != nil {
		return err
	}
	backupPath, err := ioutil.TempDir(".", "rancher-backup")
	if err != nil {
		return err
	}
	for _, crd := range CRDs.Items {
		group, version := crd.Spec.Group, crd.Spec.Versions[0].Name
		dyn, err := dynamic.NewForConfig(config)
		if err != nil {
			return err
		}
		var dr dynamic.ResourceInterface
		gvr := schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: crd.Spec.Names.Plural,
		}
		dr = dyn.Resource(gvr)
		cr, err := dr.List(v1.ListOptions{})
		if err != nil {
			return err
		}
		for _, item := range cr.Items {
			metadata := convert.ToMapInterface(item.Object["metadata"])
			delete(metadata, "creationTimestamp")
			delete(metadata, "resourceVersion")
			delete(metadata, "uid")
			item.Object["metadata"] = metadata
			writeToFile(item.Object, backupPath)
		}
	}

	return nil
}

// from velero https://github.com/vmware-tanzu/velero/blob/master/pkg/backup/item_collector.go#L267
func writeToFile(item map[string]interface{}, backupPath string) (string, error) {
	f, err := ioutil.TempFile(backupPath, "")
	if err != nil {
		return "", errors.Wrap(err, "error creating temp file")
	}
	defer f.Close()

	jsonBytes, err := json.Marshal(item)
	if err != nil {
		return "", errors.Wrap(err, "error converting item to JSON")
	}

	if _, err := f.Write(jsonBytes); err != nil {
		return "", errors.Wrap(err, "error writing JSON to file")
	}

	if err := f.Close(); err != nil {
		return "", errors.Wrap(err, "error closing file")
	}

	return f.Name(), nil
}
