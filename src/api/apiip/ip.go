package apiip

import (
	"encoding/json"
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const ApiURLQueryIpLocation = "https://kzipglobal.market.alicloudapi.com/api/ip/query"

type QueryIpLocationData struct {
	OrderNo  string `json:"orderNo"`
	Nation   string `json:"nation"`
	Province string `json:"province"`
	City     string `json:"city"`
	Ip       string `json:"ip"`
	Isp      string `json:"isp"`
}

type QueryIpLocationBody struct {
	Msg     string               `json:"msg"`
	Success bool                 `json:"success"`
	Code    int                  `json:"code"`
	Data    *QueryIpLocationData `json:"data"`
}

func QueryIpLocation(ip string) (*QueryIpLocationData, error) {
	params := url.Values{}
	params.Add("ip", ip)

	req, err := http.NewRequest("GET", ApiURLQueryIpLocation+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("APPCODE %s", config.GetConfig().API.AppCode))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res QueryIpLocationBody
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	} else if !res.Success {
		return nil, fmt.Errorf("query ip location failed (code %d): %s", res.Code, res.Msg)
	} else if res.Data == nil {
		return nil, fmt.Errorf("query ip location failed (code %d): %s, data is nil", res.Code, res.Msg)
	}

	return res.Data, nil
}

func (d *QueryIpLocationData) CheckLocation(r *config.RuleConfig) (bool, error) {
	if r.Nation != "" && d.Nation != r.Nation {
		return false, nil
	}

	if r.NationVague != "" && !strings.Contains(d.Nation, r.NationVague) {
		return false, nil
	}

	if r.Province != "" && d.Province != r.Province {
		return false, nil
	}

	if r.ProvinceVague != "" && !strings.Contains(d.Province, r.ProvinceVague) {
		return false, nil
	}

	if r.City != "" && d.City != r.City {
		return false, nil
	}

	if r.CityVague != "" && !strings.Contains(d.Province, r.CityVague) {
		return false, nil
	}

	if r.ISP != "" && d.Isp != r.ISP {
		return false, nil
	}

	if r.ISPVague != "" && !strings.Contains(d.Province, r.ISPVague) {
		return false, nil
	}

	return true, nil
}
