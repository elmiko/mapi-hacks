/*
Copyright 2023 michael mccune

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type NodeInfo struct {
	Updated bool
	Detail  string
}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	go func() {
		for range done {
			os.Exit(0)
		}
	}()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	nodemap := make(map[string]NodeInfo)
	updates := false

	fmt.Println("---")

	for {
		if updates {
			fmt.Println("---")
			updates = false
		}

		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}

		for _, n := range nodes.Items {
			ni := NodeInfo{
				Updated: true,
			}

			if b, err := json.Marshal(n); err != nil {
				log.Fatal(err)
			} else {
				ni.Detail = string(b)
			}

			if v, found := nodemap[n.Name]; found {
				if ni.Detail == v.Detail {
					ni.Updated = false
				}
			}
			nodemap[n.Name] = ni
		}

		for _, v := range nodemap {
			if v.Updated {
				fmt.Printf("%s\n", v.Detail)
				updates = true
			}
		}

		time.Sleep(250 * time.Millisecond)
	}
}
