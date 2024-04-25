package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
)

func AmazonScraper() {

	var productName string

	fmt.Print("Digite o nome do produto: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		productName = scanner.Text()
	}

	// Define a URL da busca
	url := "https://www.amazon.com.br/s?k=" + url.QueryEscape(productName)
	fmt.Println(url)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Define um user-agent para evitar problemas com a request na amazon
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	maxRetries := 50
	retryCount := 0

	for retryCount < maxRetries && res.StatusCode == 503 {
		fmt.Println("Tentativa de requisição", retryCount+1)
		retryCount++
		time.Sleep(5 * time.Second) // Aguarda 5 segundos antes de executar novamente a requisição
		res, err = client.Do(req)   // Faz uma nova requisição
		if err != nil {
			log.Fatal(err)
		}
	}

	if retryCount == maxRetries {
		log.Fatalf("Não foi possível fazer a requisição após %d tentativas", maxRetries)
	}

	// Cria um novo arquivo .xslx
	f := excelize.NewFile()
	index, err := f.NewSheet("Produtos")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Cria novas colunas na planilha
	f.SetCellValue("Produtos", "A1", "Nome do Produto")
	f.SetCellValue("Produtos", "B1", "Preço")
	f.SetCellValue("Produtos", "C1", "Imagem")
	f.SetCellValue("Produtos", "D1", "Link")

	// Seleciona a planilha ativa
	f.SetActiveSheet(index)
	// Linha para começar a gravar os dados
	row := 2

	//CARREGA O HTML UTILIZANDO O PACKAGE GOQUERY
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Cria um arquivo .txt para salvar os produtos
	// file, err := os.Create("produtos.txt")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer file.Close()
	// date := time.Now()
	// fmt.Fprintf(file, "Data e Hora da Extração: %s\n\n\n", date.Format(time.UnixDate))

	doc.Find(".s-result-item").Each(func(i int, s *goquery.Selection) {
		title := s.Find("h2").Text()
		imgSrc, _ := s.Find(".s-image").Attr("src")
		link, _ := s.Find(".a-link-normal.s-no-outline").Attr("href")
		fullPrice := s.Find("span .a-price .a-offscreen").First().Text()

		title = strings.TrimSpace(title)
		imgSrc = strings.TrimSpace(imgSrc)
		link = strings.TrimSpace(link)

		if title != "" && fullPrice != "" && link != "" {

			fmt.Printf("Produto: %s --- Preço: %s\n", title, fullPrice)

			link := "https://amazon.com.br" + link

			f.SetCellValue("Produtos", fmt.Sprintf("A%d", row), title)
			f.SetCellValue("Produtos", fmt.Sprintf("B%d", row), fullPrice)
			f.SetCellValue("Produtos", fmt.Sprintf("C%d", row), imgSrc)
			f.SetCellValue("Produtos", fmt.Sprintf("D%d", row), link)

			row++

		}
	})

	if err := f.SaveAs("ProdutosAmazon.xlsx"); err != nil {
		fmt.Println(err)
	}
}

func main() {
	AmazonScraper()
}
