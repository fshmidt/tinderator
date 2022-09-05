package main

import (
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	"database/sql"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"postgres/postgres"
	"strings"
	"sync"
	"time"
	"unicode"
)

var palka string

func findInstString(arg string) string {
	arr := strings.Split(arg, "\n")
	for _, str := range arr {
		if strings.Contains(str, "inst") {
			miniArr := strings.Split(str, ":")
			return miniArr[len(miniArr)-1]
		}
	}
	return ""
}
func check(arg1 string, arg2 string, flag int) string {
	var nick string = ""
	fmt.Println("TRIGGERED:FLAG=", flag, ", args:", arg1, arg2)
	if flag == 0 {
		if len(arg2) > 3 || arg2 == "__" || arg2 == "___" || arg2 == "____" {

			// исправление гугловского распознавание нижнего подчеривания

			switch arg2 {
			case "__":
				palka = "_"
				return ""
			case "___":
				palka = "__"
				return ""
			case "____":
				palka = "__"
				return ""
			default:
				if strings.Contains(arg2, "instagram.com/") {
					return strings.Split(arg2, "/")[len(strings.Split(arg2, "/"))-1]
				}
				if strings.Contains(arg2, "instagram") && (arg2[9] == '.' || arg2[9] == '_') {
					arg2 = arg2[9:]
				} else if strings.Contains(arg2, "instagram") {
					arg2 = arg2[10:]
				}
				if strings.Contains(arg2, "..") {
					return ""
				}
				for r, s := range arg2 {
					if unicode.Is(unicode.Latin, s) == false && unicode.IsDigit(s) == false && s != '.' && s != '_' && s != '@' {
						if r == 0 && s == '-' {
							continue
						} else {
							return ""
						}
					} else if s != '@' {
						nick = (nick + string(s))
						if shit(nick) {
							return ""
						}
					}
				}
			}
		}
	} else {
		if arg1 == "__" || arg1 == "___" || arg1 == "____" || arg2 == "__" || arg2 == "___" || arg2 == "____" {

			// исправление гугловского распознавание нижнего подчеривания

			switch arg2 {
			case "__":
				palka = "_"
				return ""
			case "___":
				palka = "__"
				return ""
			case "____":
				palka = "__"
				return ""
			}
			switch arg1 {
			case "__":
				palka = "_"
				return ""
			case "___":
				palka = "__"
				return ""
			case "____":
				palka = "__"
				return ""
			}
		} else {
			if len(check("", arg2, 0)) > len(check("", arg1, 0)) {
				//fmt.Println("LEN FROM CHECK POST, 0:", len(check("", arg2, 0)))
				//fmt.Println("LEN FROM CHECK PREV, 0:", len(check("", arg1, 0)))
				return check("", arg2, 0)
			} else {
				//fmt.Println("LEN FROM CHECK POST, 0:", len(check("", arg2, 0)))
				//fmt.Println("LEN FROM CHECK PREV, 3:", len(check("", arg1, 0)))
				return check("", arg1, 0)
			}
		}
	}
	if nick != "" {
		return palka + nick
	}
	return nick
}

func triggers(arr string) bool {

	for _, word := range Instagram {
		if arr == word {
			return true
		}
	}
	return false
}

func shit(arr string) bool {

	for _, word := range Shit {
		if arr == word {
			return true
		}
	}
	return false
}

func checkMine(nick *string, w io.Writer, file string, trig int) {
	mine := []string{"rrooddeeff", "fshmidt", "fshmidthimself", "realpartofreality", "fshmidt.store"}
	for _, word := range mine {
		if *nick == word {
			detectText(nick, w, file, trig+5)
		}
	}
}

func detectText(nick *string, w io.Writer, file string, iter int) error {
	ctx := context.Background()
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		fmt.Println(err)
		return err
	}
	f, err := os.Open(file)
	if err != nil {
		fmt.Println("2", err)
		return err
	}
	defer f.Close()

	image, err := vision.NewImageFromReader(f)
	if err != nil {
		fmt.Println("3", err)
		return err
	}
	annotations, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		return err
	}

	if len(annotations) == 0 {
		fmt.Fprintln(w, "No text found in file", file)
	} else {
		var found bool
		for i := iter; i < len(annotations); i++ {
			var arr string
			arr = annotations[i].Description
			if triggers(arr) == true {
				triggered, x := i, i

				// собираем ник из говна палок и смайликов которые обычно пишут девочки
				if i+1 < len(annotations) && (annotations[i+1].Description == ":" || annotations[i+1].Description == "-" || annotations[i+1].Description == "@") {
					x += 1
					if i+2 < len(annotations) && (annotations[i+2].Description == ":" || annotations[i+2].Description == "-" || annotations[i+2].Description == "@") {
						x += 1
					}
				}

				for n := 1; x+n < len(annotations) && i-n >= 0; n++ {
					*nick = check((annotations[i-n].Description), (annotations[x+n].Description), 1)
					if *nick != "" {
						checkMine(nick, w, file, triggered)
						break
					}
				}
				if i == len(annotations)-1 {
					for n := 1; i-n >= 0; n++ {
						*nick = check("", (annotations[i-n].Description), 0)
						if *nick != "" {
							checkMine(nick, w, file, triggered)
							break
						}
					}
				}

				if *nick != "" && *nick != "_" && len(annotations) > i+2 && (annotations[i+2].Description) == "___" && (annotations[i+2].Description) == "____" {
					*nick = *nick + "__"
				}
				if *nick != "" && *nick != "_" && i+3 < len(annotations) && ((annotations[i+3].Description) == "___" || (annotations[i+3].Description) == "____") {
					*nick = *nick + "__"
				}
				if *nick != "" && *nick != "_" {
					checkMine(nick, w, file, triggered)
					found = true
					break
				}
			}
		}
		if found == false {
			*nick = findInstString(annotations[0].Description)
		}
		//}
		if *nick == "" || found == false {
			for i, annotation := range annotations[1:] {
				i += 1
				var arr string
				arr = (annotation.Description)
				for _, y := range Instagram {
					if strings.Contains(arr, y) {
						instShit := arr
						*nick = check("", strings.Split(instShit, string(instShit[len(y)-1]))[1], 0)
						if *nick != "" {
							found = true
							break
						}
					}
				}
				if arr == "себе" && (annotations[i-1].Description) == "О" && found == false {
					for n := 1; i+n < len(annotations) && i-n-1 >= 0; n++ {
						*nick = check((annotations[i-n-1].Description), (annotations[i+n].Description), 1)
						if *nick != "" {
							found = true
							break
						}
					}
					if found == false {
						for n := len(annotations) - i - 1; i+n > i; n-- {
							*nick = check("", (annotations[i+n].Description), 0)
							if *nick != "" {
								found = true
								break
							}
						}
					}
					break
				}
			}

		}
	}
	palka = ""
	return nil
}

func parseFolder(path string) (string, error) {
	var screenshots string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		dpath, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		if strings.Contains(path, ".DS_Store") == false {
			screenshots = screenshots + "\n" + dpath + "/" + path
		}
		return nil
	})
	return screenshots, err
}

func initStore() (*sql.DB, error) {

	pgConnString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
	)

	var (
		db  *sql.DB
		err error
	)
	openDB := func() error {
		db, err = sql.Open("postgres", pgConnString)
		return err
	}

	err = backoff.Retry(openDB, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS chicks(\nUIN serial NOT NULL,\nNickname text NOT NULL,\nLiked boolean  DEFAULT FALSE,\nDM boolean DEFAULT FALSE,\nParseDate timestamp NOT NULL,\nExecDate timestamp  DEFAULT NULL,\nCONSTRAINT chickstable PRIMARY KEY (UIN)\n);"); err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	fmt.Println("STARTED AT")
	fmt.Println(time.Now().Format("Jan _2 15:04:05"))
	tm := time.Now()
	defer fmt.Println(time.Now().Format("Jan _2 15:04:05"))

	// парсим список скринов

	screenshots, err := parseFolder("/data/assets/current_batch")
	if err != nil {
		log.Fatal(err)
	}

	arr := strings.Split(screenshots, "\n")
	arr = arr[2:]

	// создаем файл для списка ников

	f, err := os.Create("/data/assets/current_list.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// открываем файл с глобальным листом ников для сверки повторов

	dataFromFile, _ := ioutil.ReadFile("/data/assets/global_list.txt")
	globalList := string(dataFromFile)

	f2, err := os.OpenFile("/data/assets/global_list.txt", os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f2.Close()

	// Открываем и прописываем закрытие базы данных

	//db := postgres.OpenDB()
	//defer postgres.CloseDB(db)
	db, err := initStore()
	if err != nil {
		log.Fatalf("failed to initialise the store: %s", err)
	}
	defer db.Close()

	// синхронизация горутин

	var new, old, a, start, end int
	var wg sync.WaitGroup
	chans := make(map[int]chan int)
	for b, _ := range arr {
		chans[b] = make(chan int, 1)
		a = b
	}
	chans[a+1] = make(chan int)
	a = len(arr)/200 + 1

	//поехали: несколько циклов по 200 горутин, чтобы гугл не сбрасывал подключение из-за долгой загрузки множества файлов

	for ; a > 0; a-- {
		if len(arr)-start > 200 {
			end = start + 200
		} else {
			end = len(arr)
		}
		for y, file := range arr[start:end] {
			wg.Add(1)
			go func(wg *sync.WaitGroup, file string, in, out chan int, y int) {
				var nick string
				fmt.Println("In gorutine ", start+y)
				err = detectText(&nick, os.Stdout, file, 0)
				if err != nil {
					panic(err)
				}
				nick = nick + "\n"
				<-in
				if strings.Contains(globalList, "\n"+nick) == false {
					fmt.Println(y+start, ": ", nick, "is taken")
					_, err = f.WriteString(nick)
					if err != nil {
						panic(err)
					}

					//mu.Lock()
					_, err = f2.WriteString(nick)
					if err != nil {
						panic(err)
					}
					postgres.ChlistToDB(nick, time.Now().Format("2006-01-02"), db)
					new++
				} else {
					fmt.Println(y, ": she's already done .... ", nick)
					_, err = f.WriteString(strings.Split(file, "/")[len(strings.Split(file, "/"))-1] + "<---------------------\n")
					if err != nil {
						panic(err)
					}
					old++
				}
				out <- 1
				wg.Done()
			}(&wg, file, chans[y], chans[y+1], y)
			if y == 0 {
				chans[y] <- 1
			}
			if y == len(arr)-1 {
				<-chans[y+1]
			}
		}

		wg.Wait()
		start += 200
	}

	fmt.Println("Finished: got ", new, "new chicks and found ", old, "old chikcs.")
	fmt.Println(time.Now().Format("Jan _2 15:04:05"))
	fmt.Println("На ", len(arr), " итераций ушло ", time.Since(tm).Round(time.Second), "минут")

}
