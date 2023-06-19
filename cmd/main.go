package main

import (
	"encoding/json"
	"fmt"
	"github.com/yuyz0112/sunmao-ui-go-binding/pkg/arco"
	"os"
	"time"

	"github.com/yuyz0112/sunmao-ui-go-binding/pkg/dovetail"
	"github.com/yuyz0112/sunmao-ui-go-binding/pkg/runtime"
	"github.com/yuyz0112/sunmao-ui-go-binding/pkg/sunmao"
)

func main() {
	// init the runtime
	r := runtime.New("ui", "patch")
	// init an App builder, use a lib
	app := sunmao.NewApp()
	b := sunmao.NewChakraUIApp()
	d := dovetail.NewDovetailApp(app)

	arcoB := arco.NewArcoApp(app)

	// build some UI in code, perfect it in the browser!
	// more UI
	app.Component(app.NewStack().Children(map[string][]sunmao.BaseComponentBuilder{
		"content": {
			app.NewText().Content("呵呵呵"),
			app.NewText().Content("哈哈哈123"),
			app.NewText().Content("Hello Sunmao!!!"),
			arcoB.NewTable(),
			arcoB.NewButton(),
		},
	}).Properties(map[string]interface{}{
		"direction": "vertical",
	}))
	app.Component(d.NewRoot().Children((map[string][]sunmao.BaseComponentBuilder{
		"root": {
			d.NewButton(),
			d.NewKubectlGetTable(),
		},
	})))

	// use server dynamic data
	entries, _ := os.ReadDir(".")
	data := []map[string]interface{}{}
	for _, e := range entries {
		info, _ := e.Info()
		data = append(data, map[string]interface{}{
			"name":    e.Name(),
			"size":    info.Size(),
			"modTime": info.ModTime().Format(time.UnixDate),
			"is_dir":  info.IsDir(),
		})
	}

	app.Component(arcoB.NewTable().Data(data).Column(&arco.ArcoTableColumn{
		DataIndex:    "name",
		Title:        "Name",
		Type:         "link",
		DisplayValue: "{{ $listItem.name }} - {{ $listItem.size }}",
	}).Column(&arco.ArcoTableColumn{
		DataIndex: "size",
		Title:     "File Size",
	}).Column(&arco.ArcoTableColumn{
		DataIndex: "modTime",
		Title:     "Modify Time",
	}).Column(&arco.ArcoTableColumn{
		Title:     "Type",
		DataIndex: "is_dir",
		Type:      "module",
		Module: &sunmao.ModuleContainer{
			Type: "custom/v1/file_type",
			Properties: map[string]any{
				"is_dir": "{{ $listItem.is_dir }}",
			},
		},
	}).Hidden("{{ tabs.activeTab != 0 }}"))

	fileTypeModule := sunmao.NewModule().
		Version("custom/v1").
		Name("file_type").
		Properties(map[string]any{
			"is_dir": "{{false}}",
		})

	fileTypeB := sunmao.NewApp()

	fileTypeModule.Impl(fileTypeB.Component(fileTypeB.NewStack().Children(map[string][]sunmao.BaseComponentBuilder{
		"content": {
			fileTypeB.NewComponent().Type("arco/v1/tag").
				Properties(map[string]interface{}{
					"content":        "file",
					"closable":       false,
					"checkable":      false,
					"defaultChecked": false,
					"color":          "",
					"size":           "small",
					"bordered":       false,
					"defaultVisible": true,
				}).Hidden("{{!is_dir}}"),
			fileTypeB.NewComponent().Type("arco/v1/tag").
				Properties(map[string]interface{}{
					"content":        "dir",
					"closable":       false,
					"checkable":      false,
					"defaultChecked": false,
					"color":          "rgba(0,180,42, 1)",
					"size":           "small",
					"bordered":       false,
					"defaultVisible": true,
				}).Hidden("{{is_dir}}"),
		},
	})).ValueOf())

	// use server push real-time data
	type MyState struct {
		Random int `json:"random"`
	}
	myState := r.NewServerState("server_push", &MyState{})
	b.Component(myState.AsComponent())

	b.Component(b.NewText().Content("data from server {{ server_push.state.random }}"))

	go func() {
		for {
			time.Sleep(1 * time.Second)
			// update state
			//myState.SetState(&MyState{Random: rand.Int()}, nil)
		}
	}()

	// add any server function as an API
	r.Handle("debug", func(m *runtime.Message, connId int) error {
		fmt.Println("debug >", m, "from >", connId)
		return nil
	})

	r.Handle("writeFile", func(m *runtime.Message, connId int) error {
		content, _ := json.Marshal(m.Params)
		err := os.WriteFile("test", content, 777)
		if err != nil {
			return err
		}

		// call any UI's method like an API
		r.Execute(&runtime.ExecuteTarget{
			Id:     "my_input",
			Method: "setInputValue",
			Parameters: map[string]interface{}{
				"value": time.Now().Format(time.UnixDate),
			},
		}, &connId)

		return nil
	})

	app.Component(b.NewButton().Content("click to debug").
		OnClick(&sunmao.ServerHandler{
			Name: "debug",
			Parameters: map[string]interface{}{
				// use ID to access component state
				"dynamic": "input value {{ my_input.value }}",
			},
		}))

	r.LoadModule(fileTypeModule)
	r.LoadApp(app)

	// start the runtime
	r.Run()
}
