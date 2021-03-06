/*
Copyright © 2021 wubin1989 <328454505@qq.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/unionj-cloud/go-doudou/esutils"
	"github.com/unionj-cloud/go-doudou/logutils"
	"github.com/unionj-cloud/go-doudou/pathutils"
	"github.com/unionj-cloud/go-doudou/svc"
)

// save generated openapi 3.0 compatible json document to elasticsearch for further use
var esaddr string
var esindex string

// publishCmd represents the http command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "index openapi 3.0 spec json file content along with some meta info to elasticsearch for further use",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		var svcdir string
		if len(args) > 0 {
			svcdir = args[0]
		}
		var err error
		if svcdir, err = pathutils.FixPath(svcdir, ""); err != nil {
			logrus.Panicln(err)
		}
		esclient, err := elastic.NewSimpleClient(
			elastic.SetErrorLog(logutils.NewLogger()),
			elastic.SetURL([]string{esaddr}...),
			elastic.SetGzip(true),
		)
		if err != nil {
			panic(fmt.Errorf("call NewSimpleClient() error: %+v\n", err))
		}
		es := esutils.NewEs(esindex, esindex, esutils.WithClient(esclient))
		s := svc.Svc{
			Dir:     svcdir,
			DocPath: docpath,
			Es:      es,
		}
		logrus.Infof("doc indexed. es doc id: %s\n", s.Publish())
	},
}

func init() {
	svcCmd.AddCommand(publishCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// publishCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	publishCmd.Flags().StringVarP(&esaddr, "esaddr", "", "", `elasticsearch instance connection address, save generated openapi 3.0 compatible json document to elasticsearch for further use`)
	publishCmd.Flags().StringVarP(&esindex, "esindex", "", "", `elasticsearch index name for saving openapi 3.0 compatible json documents`)
}
