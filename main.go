package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"

	"net/http"

	"io/ioutil"

	"strings"

	"github.com/adair/stations"
	"github.com/docopt/docopt-go"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
)

type TrainsCollecion struct {
	header           []string
	available_trains []interface{}
	options          string
}

func main() {
	usage := `
Usage:
    tickets [-gdtkz] <from> <to> <date>
Options:
    -h,--help   显示帮助菜单
    -g          高铁
    -d          动车
    -t          特快
    -k          快速
    -z          直达
Example:
    tickets 北京 上海 2016-10-10
    tickets -dg 成都 南京 2016-10-10
`
	arguments, _ := docopt.Parse(usage, nil, true, "", false)
	from_station := stations.Stations[arguments["<from>"].(string)]
	to_station := stations.Stations[arguments["<to>"].(string)]
	date := arguments["<date>"].(string)
	url := fmt.Sprintf("https://kyfw.12306.cn/otn/leftTicket/queryA?leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=ADULT", date, from_station, to_station)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	response, err := client.Get(url)
	if err != nil {
		fmt.Printf("url=%s\n, err=%s\n", url, err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("err = %s\n", err)
	}

	param_info := make(map[string]interface{})
	err = json.Unmarshal(body, &param_info)
	if err != nil {
		fmt.Printf("err = %s\n", err)
	}
	available_trains := param_info["data"].([]interface{})
	options := ""
	for i, j := range arguments {
		if j == true {
			options += i
		}
	}

	tc := &TrainsCollecion{}
	tc.initial(available_trains, options).prettyPrint()
}

func (tc *TrainsCollecion) initial(available_trains []interface{}, options string) *TrainsCollecion {
	tc.header = []string{"车次", "车站", "时间", "历时", "一等", "二等", "软卧", "硬卧", "硬座", "无座"}
	tc.available_trains = available_trains
	tc.options = options

	return tc
}

func (tc *TrainsCollecion) getDuraion(raw_train map[string]interface{}) string {
	information := raw_train["queryLeftNewDTO"].(map[string]interface{})
	duration := strings.Replace(information["lishi"].(string), ":", "小时", 1) + "分"
	return duration
}

func (tc *TrainsCollecion) trains() [][]string {
	var trains [][]string
	for _, i := range tc.available_trains {
		raw_train := i.(map[string]interface{})
		information := raw_train["queryLeftNewDTO"].(map[string]interface{})
		train_no := information["station_train_code"].(string)
		mark := strings.ToLower(train_no[0:1])
		if tc.options == "" || strings.Contains(tc.options, mark) {
			train := []string{
				train_no,
				aurora.Green(information["from_station_name"]).String() + "-" + aurora.Red(information["to_station_name"]).String(),
				aurora.Green(information["start_time"]).String() + "-" + aurora.Red(information["arrive_time"]).String(),
				tc.getDuraion(raw_train),
				information["zy_num"].(string),
				information["ze_num"].(string),
				information["rw_num"].(string),
				information["yw_num"].(string),
				information["yz_num"].(string),
				information["wz_num"].(string),
			}
			trains = append(trains, train)
		}
	}
	return trains
}

func (tc *TrainsCollecion) prettyPrint() {
	pt := tablewriter.NewWriter(os.Stdout)
	pt.SetHeader(tc.header)
	for _, v := range tc.trains() {
		pt.Append(v)
	}
	pt.Render()
}
