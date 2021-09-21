package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"sshw"

	"github.com/manifoldco/promptui"
)

const prev = "-parent-"

var (
	Version   = "20210921"
	Build     = "devel"
	BuildTime string
	GoVersion string = runtime.Version()
	V                = flag.Bool("v", false, " show version ")
	H                = flag.Bool("h", false, " show help ")
	S                = flag.Bool("s", false, " use local ssh config '~/.ssh/config' ")
	U                = flag.Bool("u", false, " ✨ Updata sshw config '(~/|./).sshw(.yaml|.yml)' ")
	Z                = flag.Bool("z", false, " ✨ Download sshw config '(~/|./).sshw(.yaml|.yml)' ")

	log = sshw.GetLogger()

	templates = &promptui.SelectTemplates{
		Label:    "✨ {{ . | green}}",
		Active:   "➤ {{ .Name | cyan  }}{{if .Alias}}({{.Alias | yellow}}){{end}} {{if .Host}}{{if .User}}{{.User | faint}}{{`@` | faint}}{{end}}{{.Host | faint}}{{end}}",
		Inactive: "  {{.Name | faint}}{{if .Alias}}({{.Alias | faint}}){{end}} {{if .Host}}{{if .User}}{{.User | faint}}{{`@` | faint}}{{end}}{{.Host | faint}}{{end}}",
	}
)

func findAlias(nodes []*sshw.Node, nodeAlias string) *sshw.Node {
	for _, node := range nodes {
		if node.Alias == nodeAlias {
			return node
		}
		if len(node.Children) > 0 {
			return findAlias(node.Children, nodeAlias)
		}
	}
	return nil
}

func main() {
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		return
	}

	if *H {
		flag.Usage()
		return
	}

	if *V {
		fmt.Println("sshw - ssh client wrapper for automatic login")
		fmt.Println("     Version::", Version)
		fmt.Println("     GitVersion::", Build)
		fmt.Println("     GoVersion ::", GoVersion)
		fmt.Println("     BuildTime ::", BuildTime)
		return
	}

	if *S {
		err := sshw.LoadSshConfig()
		if err != nil {
			log.Error("load ssh config error", err)
			os.Exit(1)
		}
	} else {
		err := sshw.LoadConfig()
		if err != nil {
			log.Error("load config error", err)
			os.Exit(1)
		}
	}
	// TODO: 新的指令
	// 1. conf:: Local To Remote
	// 2. conf:: Remote To Local
	// TODO: 210919--210922
	if *U {
		fmt.Println("sshw:: Use Upload Local Config Remote Server!! ")
		return
	}

	if *Z {
		fmt.Println("sshw:: Use Remote Config Download Local!! ")
		return
	}

	// login by alias
	if len(os.Args) > 1 {
		var nodeAlias = os.Args[1]
		var nodes = sshw.GetConfig()
		var node = findAlias(nodes, nodeAlias)
		if node != nil {
			client := sshw.NewClient(node)
			client.Login()
			return
		}
	}

	node := choose(nil, sshw.GetConfig())
	if node == nil {
		return
	}

	client := sshw.NewClient(node)
	client.Login()
}

func choose(parent, trees []*sshw.Node) *sshw.Node {
	prompt := promptui.Select{
		Label:        "select host",
		Items:        trees,
		Templates:    templates,
		Size:         20,
		HideSelected: true,
		Searcher: func(input string, index int) bool {
			node := trees[index]
			content := fmt.Sprintf("%s %s %s", node.Name, node.User, node.Host)
			if strings.Contains(input, " ") {
				for _, key := range strings.Split(input, " ") {
					key = strings.TrimSpace(key)
					if key != "" {
						if !strings.Contains(content, key) {
							return false
						}
					}
				}
				return true
			}
			if strings.Contains(content, input) {
				return true
			}
			return false
		},
	}
	index, _, err := prompt.Run()
	if err != nil {
		return nil
	}

	node := trees[index]
	if len(node.Children) > 0 {
		first := node.Children[0]
		if first.Name != prev {
			first = &sshw.Node{Name: prev}
			node.Children = append(node.Children[:0], append([]*sshw.Node{first}, node.Children...)...)
		}
		return choose(trees, node.Children)
	}

	if node.Name == prev {
		if parent == nil {
			return choose(nil, sshw.GetConfig())
		}
		return choose(nil, parent)
	}

	return node
}