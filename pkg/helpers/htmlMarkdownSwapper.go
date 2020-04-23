package helpers

import (
	"github.com/gomarkdown/markdown"
)

func ConvertHtmlToMarkdown(html string) string{
	// conv := md.NewConverter("", true, nil)
	// // Use the `GitHubFlavored` plugin from the `plugin` package.
	// conv.Use(plugin.GitHubFlavored())
	//
	// markDownConv, err := conv.ConvertString(html)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// return markDownConv

	return html
}
func ConvertMarkdownToHtml(markdownStr string) string {
	output := markdown.ToHTML([]byte(markdownStr), nil, nil)
	return string(output)
}
