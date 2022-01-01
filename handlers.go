package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"os"
	"unicode/utf8"

	"github.com/joho/godotenv"
	"github.com/yanzay/tbot/v2"
	"github.com/PuerkitoBio/goquery"
	ocr "github.com/ranghetto/go_ocr_space"
	)

var ocrtoken string

func init(){
	e := godotenv.Load()
	if e != nil {
		log.Println(e)
	}
	ocrtoken = os.Getenv("OCR_TOKEN")
}

// Handle the /start command here
func (a *application) startHandler(m *tbot.Message) {
	msg := "\n*Привет!* Присылай сообщение и я его переведу.\n\nДля обычного текста:\n*en* *ru* Hello, World!" +
	       "\nДля текста на фотографии:\n*en* *ru* https://website.com/file\n\nГде *en* - это язык текста, " +
	       "а *ru* - язык перевода.\nДля *OCR* без перевода используй одинаковые языки.\n\nКоды языков:\n*ru* - Русский; " +
	       "*en* - Английский; *ar* - Арабский; *zh* - Китайский; \n*fr* - Французкий; *de* - Немецкий; *hi* - Индийский; " +
	       "\n*id* - Индонезийский; *ga* - Ирландский; *it* - Итальянский; \n*ja* - Японский; *ko* - Корейский; " +
	       "*pl* - Польский; *pt* - Португальский; \n*es* - Испанский; *tr* - Турецкий; *vi* - Вьетнамский."
	a.client.SendMessage(m.Chat.ID, msg, tbot.OptParseModeMarkdown)
}

// Handle the msg command here
func (a *application) msgHandler(m *tbot.Message) {
	msg := ""
	source := ""
	target := ""
	text := ""
	languages := map[string]string{
		"ru": "rus",
		"en": "eng",
		"ar": "ara",
		"zh": "chs",
		"fr": "fre",
		"de": "ger",
		"it": "kor",
		"ja": "jpn",
		"pl": "pol",
		"pt": "por",
		"tr": "tur",
		"hi": "Распознование хинди не поддерживается.",
		"id": "Распознование индонезийского не поддерживается.",
		"ga": "Распознование ирландского не поддерживается.",
		"es": "Распознование испанского не поддерживается.",
		"vi": "Распознование вьетнамского не поддерживается.",
	}
	if utf8.RuneCountInString(m.Text) > 6{
		arr := strings.Split(m.Text, " ")
		source = arr[0]
		target = arr[1]
		text = strings.TrimLeft(m.Text, arr[0]+arr[1])
		if languages[source] == ""{
			msg = "Неправильный код языка текста!\nПосмотри коды командой */start*."
		} else if languages[target] == ""{
			msg = "Неправильный код языка перевода!\nПосмотри коды командой */start*."
		}
	}else{
		msg = "Слишком короткое сообщение!\nПосмотри на пример командой */start*."}
	if msg == "" && utf8.RuneCountInString(m.Text) > 15{
		if text[:len(text)-(len(text)-4)] == "http" {
			res, err := http.Get("https://status.ocr.space/")
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()
			if res.StatusCode != 200 {
				log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
			}
			doc, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			doc.Find("span").Each(func(i int, s *goquery.Selection) {
				fruw := ""
				if i == 0 {
					fruw = strings.Trim(s.Text(), " ")
				if fruw != "UP" {
					msg = "В данный момент сервера OCR недоступны."
						}
			}
			})}
			if msg == "" && text[:len(text)-(len(text)-4)] == "http"{
			//токен + язык
			config := ocr.InitConfig(ocrtoken, languages[source])
			//урл
			result, err := config.ParseFromUrl(text)
			if err != nil {
				fmt.Println(err)
				}
				//вывод текста
				text = fmt.Sprintln(result.JustText())
			}
		}
	if msg == ""{
		message := map[string]interface{}{
		"q":      text,
		"source": source,
		"target": target,
		}
		bytesRepresentation, err := json.Marshal(message)
		if err != nil {
			log.Println(err)
		}

		resp, err := http.Post("https://libretranslate.de/translate", "application/json",
			bytes.NewBuffer(bytesRepresentation))
		if err != nil {
			log.Println(err)
		}
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		msg = fmt.Sprintln(result)
		msg = strings.Trim(msg, "map[translatedText:")
		msg = msg[:len(msg)-2]
		}		
	a.client.SendMessage(m.Chat.ID, msg, tbot.OptParseModeMarkdown)
	}
