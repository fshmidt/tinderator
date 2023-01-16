package main

import (
	vision "cloud.google.com/go/vision/apiv1"
	"context"
	"database/sql"
	"fmt"
	backoff "github.com/cenkalti/backoff/v4"
	"github.com/joho/godotenv"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"postgres/postgres"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

func goodSymbosl(str string) bool {

	for _, s := range str {
		if unicode.Is(unicode.Latin, s) == false && unicode.IsDigit(s) == false && s != '.' && s != '_' && s != '@' {
			//fmt.Println("FALSE!")
			return false
		}
	}
	return true
}

func findInstString(arg string) string {
	arr := strings.Split(arg, "\n")
	maxLen := 0
	var candidate string
	for i, str := range arr {
		for _, word := range Instagram {
			if str == word {
				miniArr := str[len(word):]
				for x, y := range miniArr {
					if unicode.Is(unicode.Latin, y) || unicode.IsDigit(y) || y == '.' || y == '_' {
						return miniArr[x:]
					}
				}
			}
		}
		if str == "@" && check(arr[i+1]) != "" {
			//fmt.Println("FINDSTRING: CHECKING", arr[i+1], "RETURNED:", check(arr[i+1]))
			return check(arr[i+1])
		} else if str[0] == '@' && check(str[1:]) != "" {
			return str[1:]
		}
		if strings.Contains(str, " ") {
			subArr := strings.Split(str, " ")
			for _, subStr := range subArr {
				if len(subStr) > maxLen && goodSymbosl(subStr) && !shit(subStr) {
					candidate = subStr
					maxLen = len(candidate)
				}
			}
		} else if len(str) > maxLen && goodSymbosl(str) && !shit(str) {
			candidate = str
			maxLen = len(candidate)
		}

	}
	if len(candidate) > 1 && candidate[0] == '@' {
		return candidate[1:]
	}
	return candidate
}
func check(arg string) string {
	var nick string = ""
	if len(arg) > 3 {
		if strings.Contains(arg, "instagram.com/") {
			return strings.Split(arg, "/")[len(strings.Split(arg, "/"))-1]
		}
		if strings.Contains(arg, "instagram") && len(arg) > 9 && (arg[9] == '.' || arg[9] == '_' || arg[9] == ':') {
			arg = arg[9:]
		} else if strings.Contains(arg, "instagram") && len(arg) > 9 {
			arg = arg[10:]
		}
		for r, s := range arg {
			if unicode.Is(unicode.Latin, s) == false && unicode.IsDigit(s) == false && s != '.' && s != '_' && s != '@' {
				if r == 0 && s == '-' {
					continue
				} else if len(nick) >= 3 {
					return ("WRONG_SYMB-" + strconv.Itoa(r))
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
	return nick
}

func triggers(arr string) bool {

	for _, word := range Instagram {
		if word == arr {
			return true
		}
	}
	return false
}

func dogHunter(arr string) bool {
	if arr[0] == '@' && len(arr) > 4 {
		//fmt.Println("WOOF WOOWF>>>>>>>>>>>>>>>>>")
		return true
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
		//fmt.Println("3", err)
		return err
	}
	annotations, err := client.DetectTexts(ctx, image, nil, 10)
	if err != nil {
		return err
	}

	candidate := ""
	if len(annotations) == 0 {
		fmt.Fprintln(w, "No text found in file", file)
	} else {
		found := false
		for i := iter; i < len(annotations); i++ {
			var arr string
			arr = annotations[i].Description
			if dogHunter(arr) == true {
				candidate = check(arr)
				if candidate != "" {
					checkMine(&candidate, w, file, i)
				}
			}
			if triggers(arr) == true && i > 1 && annotations[i-1].Description != "Tele2" {
				triggered, x := i, i
				if i+1 < len(annotations) && (annotations[i+1].Description == ":" || annotations[i+1].Description == "-" || annotations[i+1].Description == "@") {
					x += 1
					if i+2 < len(annotations) && (annotations[i+2].Description == ":" || annotations[i+2].Description == "-" || annotations[i+2].Description == "@") {
						x += 1
					}
				}
				prev := ""
				for n := 0; x+n < len(annotations); n++ {

					*nick = check(annotations[x+n].Description)
					if *nick != "" {
						checkMine(nick, w, file, triggered)
						break
					}
					if n <= 2 {
						prev = check(annotations[x-n].Description)
						if prev != "" {
							*nick = prev
							return nil
						}
					}
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
		if *nick == "" {
			*nick = candidate
		}
	}
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

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	// парсим список скринов

	screenshots, err := parseFolder("../assets/current_batch")
	if err != nil {
		log.Fatal(err)
	}

	arr := strings.Split(screenshots, "\n")
	arr = arr[2:]

	// создаем файл для списка ников

	f, err := os.Create("../assets/current_list.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// открываем файл с глобальным листом ников для сверки повторов

	dataFromFile, _ := ioutil.ReadFile("../assets/global_list.txt")
	globalList := string(dataFromFile)

	f2, err := os.OpenFile("../assets/global_list.txt", os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f2.Close()

	// Открываем и прописываем закрытие Postgres базы данных

	db := postgres.OpenDB()
	defer postgres.CloseDB(db)
	db, err = initStore()
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
	a = len(arr)/280 + 1

	//поехали: несколько циклов по 280 горутин, чтобы гугл не сбрасывал подключение из-за долгой загрузки множества файлов

	for ; a > 0; a-- {
		if len(arr)-start > 280 {
			end = start + 280
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
					fmt.Println("Panic at", file)
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

					_, err = f2.WriteString(nick)
					if err != nil {
						panic(err)
					}
					//postgres.ChlistToDB(nick, time.Now().Format("2006-01-02"), db)
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
		start += 280
	}

	fmt.Println("Finished: got ", new, "new chicks and found ", old, "old chikcs.")
	fmt.Println(time.Now().Format("Jan _2 15:04:05"))
	fmt.Println("На ", len(arr), " итераций ушло ", time.Since(tm).Round(time.Second))

}
