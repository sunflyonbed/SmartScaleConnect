package picooc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AlexxIT/SmartScaleConnect/pkg/core"
)

type Client struct {
	client *http.Client

	deviceID string
	token    string
	userID   string

	roleIDs map[string]string
	roles   []roleInfo
}

type roleInfo struct {
	ID   string
	Name string
}

type bodyRecord struct {
	BodyTime         int64   `json:"bodyTime"`
	BodyFat          float32 `json:"body_fat"`
	Weight           float32 `json:"weight"`
	BMI              float32 `json:"bmi"`
	VisceralFatLevel float32 `json:"visceral_fat_level"`
	BodyAge          float32 `json:"body_age"`
	BoneMass         float32 `json:"bone_mass"`
	BasicMetabolism  float32 `json:"basic_metabolism"`
	WaterRace        float32 `json:"water_race"`
	SkeletalMuscle   float32 `json:"skeletal_muscle"`
	IsDel            int     `json:"is_del"`
	AbnormalFlag     int     `json:"abnormal_flag"`
	MAC              string  `json:"mac"`
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{Timeout: time.Minute},
	}
}

func (c *Client) GetAllWeights() ([]*core.Weight, error) {
	return c.GetFilterWeights("")
}

func (c *Client) GetFilterWeights(name string) ([]*core.Weight, error) {
	role, err := c.selectRole(name)
	if err != nil {
		return nil, err
	}
	log.Printf("picooc: use roleid=%s rolename=%s\n", role.ID, role.Name)

	var weights []*core.Weight

	bodyTime := time.Now().Unix()

	for {
		params := c.loadingBodyDataParams(role.ID, bodyTime)
		req, err := http.NewRequest("GET", api+"v1/api/mixData/loadingBodyData?"+params.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "okhttp/3.14.7")

		res, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}

		var res1 struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Resp struct {
				DescResult struct {
					MixDataList        []bodyRecord `json:"mixDataList"`
					AllLeftCount       int          `json:"allLeftCount"`
					BodyIndexLeftCount int          `json:"bodyIndexLeftCount"`
				} `json:"descResult"`
			} `json:"resp"`
		}

		err = json.NewDecoder(res.Body).Decode(&res1)
		res.Body.Close()
		if err != nil {
			return nil, err
		}
		if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
			return nil, errors.New("picooc: loading body data error: " + res.Status)
		}
		if res1.Code != http.StatusOK {
			return nil, errors.New("picooc: loading body data error: " + res1.Msg)
		}

		records := res1.Resp.DescResult.MixDataList
		for _, v1 := range records {
			if v1.AbnormalFlag != 0 || v1.IsDel != 0 {
				continue
			}

			w := &core.Weight{
				Date:   time.Unix(v1.BodyTime, 0),
				Weight: v1.Weight,

				BMI:       v1.BMI,
				BodyFat:   v1.BodyFat,
				BodyWater: v1.WaterRace,
				BoneMass:  v1.BoneMass,

				MetabolicAge: int(v1.BodyAge), // 0
				VisceralFat:  int(v1.VisceralFatLevel),

				BasalMetabolism:    int(v1.BasicMetabolism),
				SkeletalMuscleMass: v1.SkeletalMuscle, // 0

				User:   name,
				Source: v1.MAC,
			}
			weights = append(weights, w)
		}

		if len(records) == 0 || res1.Resp.DescResult.AllLeftCount == 0 && res1.Resp.DescResult.BodyIndexLeftCount == 0 {
			break
		}

		bodyTime = records[len(records)-1].BodyTime
	}

	return weights, nil
}

func (c *Client) loadingBodyDataParams(roleID string, bodyTime int64) url.Values {
	const (
		method     = "mixData/loadingBodyData"
		appVersion = "4.13.0"
	)

	params := c.valuesWithAppVer(method, appVersion)
	params.Set("os", "android")
	params.Set("pageSize", "200")
	params.Set("timeZone", "Asia/Shanghai")
	params.Set("userId", c.userID)
	params.Set("roleId", roleID)
	params.Set("version", "4.16.0")
	params.Set("isMainRole", strconv.FormatBool(len(c.roles) > 0 && roleID == c.roles[0].ID))
	params.Set("device_mac", c.deviceID)
	params.Set("mainRoleId", "0")
	params.Set("bodyTime", strconv.FormatInt(bodyTime, 10))
	params.Set("lastXCXBodyIndexTime", "0")
	params.Set("reqType", "2")
	params.Set("lang", "zh_CN")
	params.Set("lastDeleteBodyIndexTime", "0")
	params.Del("timezone")

	return params
}

func (c *Client) selectRole(name string) (roleInfo, error) {
	if len(c.roles) == 0 {
		return roleInfo{}, errors.New("picooc: no roles found")
	}

	if name == "" {
		return c.roles[0], nil
	}

	if roleID, ok := c.roleIDs[name]; ok {
		return roleInfo{ID: roleID, Name: name}, nil
	}

	return roleInfo{}, fmt.Errorf("picooc: unknown rolename %q, available: %s", name, c.roleNames())
}

func (c *Client) roleNames() string {
	names := make([]string, 0, len(c.roles))
	for _, role := range c.roles {
		if role.Name == "" {
			continue
		}
		names = append(names, role.Name)
	}
	return strings.Join(names, ", ")
}

func (c *Client) GetMeasureList(startTime, endTime int64, page, size int) (json.RawMessage, error) {
	return c.GetMeasureListWithAuth(c.token, c.userID, startTime, endTime, page, size)
}

func (c *Client) GetMeasureListWithAuth(token, uid string, startTime, endTime int64, page, size int) (json.RawMessage, error) {
	params, err := measureParams(token, uid)
	if err != nil {
		return nil, err
	}

	params.Set("start_time", strconv.FormatInt(startTime, 10))
	params.Set("end_time", strconv.FormatInt(endTime, 10))
	params.Set("page", strconv.Itoa(page))
	params.Set("size", strconv.Itoa(size))

	return c.getMeasure("list", params)
}

func (c *Client) GetMeasureToday() (json.RawMessage, error) {
	return c.GetMeasureTodayWithAuth(c.token, c.userID)
}

func (c *Client) GetMeasureTodayWithAuth(token, uid string) (json.RawMessage, error) {
	params, err := measureParams(token, uid)
	if err != nil {
		return nil, err
	}

	return c.getMeasure("today", params)
}

func (c *Client) GetMeasureTrend(trendType int) (json.RawMessage, error) {
	return c.GetMeasureTrendWithAuth(c.token, c.userID, trendType)
}

func (c *Client) GetMeasureTrendWithAuth(token, uid string, trendType int) (json.RawMessage, error) {
	params, err := measureParams(token, uid)
	if err != nil {
		return nil, err
	}

	params.Set("type", strconv.Itoa(trendType))

	return c.getMeasure("trend", params)
}

func measureParams(token, uid string) (url.Values, error) {
	if token == "" {
		return nil, errors.New("picooc: empty token")
	}
	if uid == "" {
		return nil, errors.New("picooc: empty uid")
	}

	return url.Values{
		"token": {token},
		"uid":   {uid},
	}, nil
}

func (c *Client) getMeasure(method string, params url.Values) (json.RawMessage, error) {
	res, err := c.client.Get(api + "v2/measure/" + method + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, errors.New("picooc: measure " + method + " error: " + res.Status)
	}

	return json.RawMessage(data), nil
}
