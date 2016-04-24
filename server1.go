package main

import (
	"encoding/json"
	"fmt"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

var rassecond = 1

type Danshi struct {
	Field     int  `json:"id"`
	Status    bool `json:"open"`
	PostDate  time.Time
	Statetime int
}

type Danshi_bson struct {
	ID        bson.ObjectId `bson:"_id`
	Field     int           `bson:"field"`
	Status    bool          `bson:"status"`
	PostDate  time.Time     `bson:"postdate"`
	Statetime int           `bson:"statetime"`
}

type Danshi_status struct {
	Status_11 bool        `json:"status_11"`
	Status_21 bool        `json:"status_21"`
	Status_22 bool        `json:"status_22"`
	Status_31 bool        `json:"status_31"`
	Status_32 bool        `json:"status_32"`
	Status_41 bool        `json:"status_41"`
	Status_42 bool        `json:"status_42"`
	Status_51 bool        `json:"status_51"`
	Status_52 bool        `json:"status_52"`
	Status_61 bool        `json:"status_61"`
	Status_62 bool        `json:"status_62"`
	Status_71 bool        `json:"status_71"`
	Status_72 bool        `json:"status_72"`
	Status_81 bool        `json:"status_81"`
	Status_82 bool        `json:"status_82"`
	Status_91 bool        `json:"status_91"`
	Status_92 bool        `json:"status_92"`
	Summary   []Summary_1 `json:"[]summary"`
}

type Summary_1 struct {
	Date     string
	Area_11  string
	Area_21  string
	Area_22  string
	Area_31  string
	Area_32  string
	Area_41  string
	Area_42  string
	Area_51  string
	Area_52  string
	Area_61  string
	Area_62  string
	Area_71  string
	Area_72  string
	Area_81  string
	Area_82  string
	Area_91  string
	Area_92  string
	Count_11 int
	Count_21 int
	Count_22 int
	Count_31 int
	Count_32 int
	Count_41 int
	Count_42 int
	Count_51 int
	Count_52 int
	Count_61 int
	Count_62 int
	Count_71 int
	Count_72 int
	Count_81 int
	Count_82 int
	Count_91 int
	Count_92 int
}

type Danshi_bson_summary struct {
	Count  int  `json:"count"`
	Year   int  `json:"year"`
	Day    int  `json:"day"`
	Month  int  `json:"month"`
	Field  int  `json:"field"`
	Status bool `json:"status"`
}

// テンプレートのコンパイル
var h = template.Must(template.ParseFiles("honoka/bootstrap.html"))

func HonokaHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	session, err := mgo.Dial("mongodb://localhost/toilet")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	db := session.DB("toilet")

	if r.Method == "POST" {
		// リクエストボディをJSONに変換
		var danshi Danshi
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&danshi)
		if err != nil {
			log.Fatal(err)
		}

		nStatus := getStatus(db, danshi.Field)
		if !danshi.Status || nStatus.Status {
			ritsu := &Danshi{danshi.Field, danshi.Status, time.Now(), 0}
			db.C("danshi").Insert(ritsu)

		} else {
			time.Since(nStatus.PostDate)
			fmt.Println("time-b")
			fmt.Println(time.Since(nStatus.PostDate))
			ritsu := &Danshi{danshi.Field, danshi.Status, time.Now(), int(time.Since(nStatus.PostDate).Seconds())}
			db.C("danshi").Insert(ritsu)
		}

		// レスポンスとしてステータスコード201を送信
		w.WriteHeader(http.StatusCreated)
	} else if r.Method == "GET" {
		p := Danshi_status{
			Status_11: getStatus(db, 11).Status,
			Status_21: getStatus(db, 21).Status,
			Status_22: getStatus(db, 22).Status,
			Status_31: getStatus(db, 31).Status,
			Status_32: getStatus(db, 32).Status,
			Status_41: getStatus(db, 41).Status,
			Status_42: getStatus(db, 42).Status,
			Status_51: getStatus(db, 51).Status,
			Status_52: getStatus(db, 52).Status,
			Status_61: getStatus(db, 61).Status,
			Status_62: getStatus(db, 62).Status,
			Status_71: getStatus(db, 71).Status,
			Status_72: getStatus(db, 72).Status,
			Status_81: getStatus(db, 81).Status,
			Status_82: getStatus(db, 82).Status,
			Status_91: getStatus(db, 91).Status,
			Status_92: getStatus(db, 92).Status,
			Summary:   select_map(getAllStatus(db)),
		}

		fmt.Println(p)

		// レスポンスにエンコーディングしたHTMLを書き込む
		h.Execute(w, p)
	}
}

func getStatus(db *mgo.Database, field int) *Danshi_bson {
	p := new(Danshi_bson)
	query := db.C("danshi").Find(bson.M{"field": field}).Sort("-postdate")
	query.One(&p)
	return p
}

func getAllStatus(db *mgo.Database) []Danshi_bson_summary {

	query := db.C("danshi").Pipe([]bson.M{
		bson.M{
			"$group": bson.M{
				"count": bson.M{"$sum": "$statetime"},
				"_id": bson.M{
					"year":   bson.M{"$year": "$postdate"},
					"month":  bson.M{"$month": "$postdate"},
					"day":    bson.M{"$dayOfMonth": "$postdate"},
					"field":  "$field",
					"status": "$status"},
			},
		},
		bson.M{
			"$project": bson.M{
				"year":   "$_id.year",
				"month":  "$_id.month",
				"day":    "$_id.day",
				"field":  "$_id.field",
				"status": "$_id.status",
				"count":  1,
			},
		},
		bson.M{
			"$sort": bson.M{
				"year":   1,
				"month":  1,
				"day":    1,
				"field":  1,
				"status": 1,
			},
		},
	})

	results := []Danshi_bson_summary{}

	err1 := query.All(&results)

	if err1 != nil {
		fmt.Printf("ERROR : %s\n", err1.Error())
	}
	fmt.Println(results)
	return results

}

func select_map(s []Danshi_bson_summary) []Summary_1 {
	ans := make([]Summary_1, 0)
	var datetmp string
	var datetmps []string
	for _, x := range s {

		datetmp = strconv.Itoa(x.Year) + strconv.Itoa(x.Month) + strconv.Itoa(x.Day)
		var zflg bool = true
		for _, d := range datetmps {
			if d == datetmp {
				zflg = false
			}
		}
		if zflg {

			datetmps = append(datetmps, datetmp)
			var Count_11_tmp int = 0
			var Count_21_tmp int = 0
			var Count_22_tmp int = 0
			var Count_31_tmp int = 0
			var Count_32_tmp int = 0
			var Count_41_tmp int = 0
			var Count_42_tmp int = 0
			var Count_51_tmp int = 0
			var Count_52_tmp int = 0
			var Count_61_tmp int = 0
			var Count_62_tmp int = 0
			var Count_71_tmp int = 0
			var Count_72_tmp int = 0
			var Count_81_tmp int = 0
			var Count_82_tmp int = 0
			var Count_91_tmp int = 0
			var Count_92_tmp int = 0

			for _, x1 := range s {
				if datetmp == strconv.Itoa(x1.Year)+strconv.Itoa(x1.Month)+strconv.Itoa(x1.Day) {
					fmt.Println(x1.Field)
					fmt.Println(x1.Count)
					if 11 == x1.Field {
						Count_11_tmp = x1.Count * rassecond
					} else if 21 == x1.Field {
						Count_21_tmp = x1.Count * rassecond
					} else if 22 == x1.Field {
						Count_22_tmp = x1.Count * rassecond
					} else if 31 == x1.Field {
						Count_31_tmp = x1.Count * rassecond
					} else if 32 == x1.Field {
						Count_32_tmp = x1.Count * rassecond
					} else if 41 == x1.Field {
						Count_41_tmp = x1.Count * rassecond
					} else if 42 == x1.Field {
						Count_42_tmp = x1.Count * rassecond
					} else if 51 == x1.Field {
						Count_51_tmp = x1.Count * rassecond
					} else if 52 == x1.Field {
						Count_52_tmp = x1.Count * rassecond
					} else if 61 == x1.Field {
						Count_61_tmp = x1.Count * rassecond
					} else if 62 == x1.Field {
						Count_62_tmp = x1.Count * rassecond
					} else if 71 == x1.Field {
						Count_71_tmp = x1.Count * rassecond
					} else if 72 == x1.Field {
						Count_72_tmp = x1.Count * rassecond
					} else if 81 == x1.Field {
						Count_81_tmp = x1.Count * rassecond
					} else if 82 == x1.Field {
						Count_82_tmp = x1.Count * rassecond
					} else if 91 == x1.Field {
						Count_91_tmp = x1.Count * rassecond
					} else if 92 == x1.Field {
						Count_92_tmp = x1.Count * rassecond
					}
				}
			}
			ans = append(
				ans,
				Summary_1{
					Date:     datetmp,
					Area_11:  "11",
					Area_21:  "21",
					Area_22:  "22",
					Area_31:  "31",
					Area_32:  "32",
					Area_41:  "41",
					Area_42:  "42",
					Area_51:  "51",
					Area_52:  "52",
					Area_61:  "61",
					Area_62:  "62",
					Area_71:  "71",
					Area_72:  "72",
					Area_81:  "81",
					Area_82:  "82",
					Area_91:  "91",
					Area_92:  "92",
					Count_11: Count_11_tmp,
					Count_21: Count_21_tmp,
					Count_22: Count_22_tmp,
					Count_31: Count_31_tmp,
					Count_32: Count_32_tmp,
					Count_41: Count_41_tmp,
					Count_42: Count_42_tmp,
					Count_51: Count_51_tmp,
					Count_52: Count_52_tmp,
					Count_61: Count_61_tmp,
					Count_62: Count_62_tmp,
					Count_71: Count_71_tmp,
					Count_72: Count_72_tmp,
					Count_81: Count_81_tmp,
					Count_82: Count_82_tmp,
					Count_91: Count_91_tmp,
					Count_92: Count_92_tmp,
				},
			)
		}
	}
	return ans
}

func main() {
	cssHandler := http.FileServer(http.Dir("./honoka/css/"))
	fontsHandler := http.FileServer(http.Dir("./honoka/fonts/"))
	jsHandler := http.FileServer(http.Dir("./honoka/js/"))
	imageHandler := http.FileServer(http.Dir("./honoka/images/"))

	http.Handle("/css/", http.StripPrefix("/css/", cssHandler))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", fontsHandler))
	http.Handle("/js/", http.StripPrefix("/js/", jsHandler))
	http.Handle("/images/", http.StripPrefix("/images/", imageHandler))

	http.HandleFunc("/toilet", HonokaHandler)
	http.ListenAndServe(":3000", nil)
}
