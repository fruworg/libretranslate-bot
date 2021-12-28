package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	ocr "github.com/ranghetto/go_ocr_space"
)

func main() {
	// токен
	bot, err := tgbotapi.NewBotAPI("telegramToken")
	if err != nil {
		log.Fatal(err)
	}
 
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	
	adrheroku := "https://libretranslate-bot.herokuapp.com:"
	wh, _ := tgbotapi.NewWebhook(adrheroku)

	_, err = bot.SetWebhook(wh)
	if err != nil {
		log.Fatal(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)

	// обрабатываем
	for update := range updates {
		if update.Message == nil {
			continue
		}
		lgocr := ""
		text := update.Message.Text
		// ответ
		reply := "Ты чё сделал?"
		switch text {
		case "/start":
			reply = "\n*Привет!* Присылай сообщение и я его переведу.\n\nДля обычного текста:\n*en**ru*Hello, World!" +
				"\nДля текста на фотографии:\n*en**ru*https://website.com/file\n\nГде *en* - это язык текста, " +
				"а *ru* - язык перевода.\nДля *OCR* без перевода используй одинаковые языки.\n\nКоды языков:\n*ru* - Русский;" +
				"*en* - Английский; *ar* - Арабский; *zh* - Китайский; \n*fr* - Французкий; *de* - Немецкий; *hi* - Индийский; " +
				"\n*id* - Индонезийский; *ga* - Ирландский; *it* - Итальянский; \n*ja* - Японский; *ko* - Корейский; " +
				"*pl* - Польский; *pt* - Португальский; \n*es* - Испанский; *tr* - Турецкий; *vi* - Вьетнамский."
		case "/help":
			reply = "нюхай бебру (*/start*)"
		default:
			if len(text) > 4 {
				source := text[:len(text)-(len(text)-2)]
				if source == "ru" || source == "en" || source == "ar" || source == "zh" || source == "fr" ||
					source == "de" || source == "hi" || source == "id" || source == "ga" || source == "it" ||
					source == "ja" || source == "ko" || source == "pl" || source == "pt" || source == "es" ||
					source == "tr" || source == "vi" {
					target := text[:len(text)-(len(text)-4)]
					target = target[len(target)-2:]
					if target == "ru" || target == "en" || target == "ar" || target == "zh" || target == "fr" ||
						target == "de" || target == "hi" || target == "id" || target == "ga" || target == "it" ||
						target == "ja" || target == "ko" || target == "pl" || target == "pt" || target == "es" ||
						target == "tr" || target == "vi" {
						text = text[len(text)-(len(text)-4):]
						//ocr
						if len(text) > 12 {
							ocryes := text[:len(text)-(len(text)-8)]
							if ocryes[:len(ocryes)-4] == "http" {
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
									if i == 0 {
										reply = strings.Trim(s.Text(), " ")
										if reply != "UP" {
											reply = "В данный момент сервера OCR недоступны."
											lgocr = "err"
										}
									}
								})
								if lgocr != "err" {
									switch source {
									case "ru":
										lgocr = "rus"
									case "en":
										lgocr = "eng"
									case "ar":
										lgocr = "ara"
									case "zh":
										lgocr = "chs"
									case "fr":
										lgocr = "fre"
									case "de":
										lgocr = "ger"
									case "hi":
										reply = "Распознование хинди не поддерживается."
										lgocr = "err"
									case "id":
										reply = "Распознование индонезийского не поддерживается."
										lgocr = "err"
									case "ga":
										reply = "Распознование ирландского не поддерживается."
										lgocr = "err"
									case "it":
										lgocr = "ita"
									case "ko":
										lgocr = "kor"
									case "ja":
										lgocr = "jpn"
									case "pl":
										lgocr = "pol"
									case "pt":
										lgocr = "por"
									case "es":
										reply = "Распознование испанского не поддерживается."
										lgocr = "err"
									case "tr":
										lgocr = "tur"
									case "vi":
										lgocr = "Распознование вьетнамского не поддерживается."
										lgocr = "err"
									default:
										reply = "Неправильный язык текста"
										lgocr = "err"
									}
									if lgocr != "err" {
										//токен + язык
										config := ocr.InitConfig("ocrToken", lgocr)
										//урл
										result, err := config.ParseFromUrl(text)
										if err != nil {
											fmt.Println(err)
										}
										//вывод текста
										text = fmt.Sprintln(result.JustText())
									}
								}
							}
						}
						if lgocr != "err" {
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
							reply = fmt.Sprintln(result)
							reply = strings.Trim(reply, "map[translatedText:")
							reply = reply[:len(reply)-2]
						}
					} else if lgocr != "err" {
						reply = "Неправильный код языка перевода!\nПосмотри коды командой */start*."
					}
				} else if lgocr != "err" {
					reply = "Неправильный код языка текста!\nПосмотри коды командой */start*."
				}
			} else if lgocr != "err" {
				reply = "Слишком короткое сообщение!\nПосмотри на пример командой */start*."
			}
		}

		// лог
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// создание ответа
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		// отправление
		msg.ParseMode = "markdown"
		bot.Send(msg)

	}
}
