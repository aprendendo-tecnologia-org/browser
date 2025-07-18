# browser
A Go package that provides a wrapper for headless browser interactions
# browser

O pacote `browser` fornece uma abstração de alto nível para automação de navegação web utilizando o [chromedp](https://github.com/chromedp/chromedp) em Go. Ele permite visitar páginas, extrair documentos HTML, interagir com elementos do DOM, aguardar elementos específicos, clicar em elementos e executar funções personalizadas sobre nós do DOM.

## Propósito

O objetivo deste pacote é facilitar a automação de tarefas em navegadores, como scraping, testes automatizados de interfaces web e manipulação de páginas dinâmicas, sem a necessidade de lidar diretamente com a API detalhada do chromedp.

## Principais Funcionalidades

- **Visitar URLs**: Navega até uma URL e armazena o documento HTML retornado.
- **Obter documento HTML**: Recupera o HTML da página atual.
- **Obter URL atual**: Retorna a URL da página atualmente carregada.
- **Aguardar elementos**: Espera até que um seletor CSS esteja visível na página.
- **Obter nós do DOM**: Extrai nós do DOM que correspondem a um seletor.
- **Clicar em elementos**: Realiza cliques em elementos da página.
- **Executar funções customizadas**: Permite executar funções sobre nós do DOM encontrados.
- **Configurar timeout**: Permite definir o tempo máximo de espera para operações do navegador.

## Exemplo de Uso

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/chromedp/chromedp"
    "seuprojeto/browser"
)

func main() {
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()

    b := browser.NewBrowser()

    status, err := b.Visit(ctx, "https://exemplo.com")
    if err != nil {
        log.Fatalf("Erro ao visitar página: %v", err)
    }
    fmt.Printf("Status: %d\n", status)
    fmt.Println("HTML:", b.GetDocument())
}