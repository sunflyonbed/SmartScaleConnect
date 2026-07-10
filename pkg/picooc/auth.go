package picooc

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const api = "https://api2.picooc.com/"

type loginResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Method    string `json:"method"`
	Token     string `json:"token"`
	UserToken string `json:"user_token"`
	UID       string `json:"uid"`
	UserID    string `json:"user_id"`
	Result    struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"result"`
	Resp loginResp `json:"resp"`
	Data struct {
		UserID    string `json:"user_id"`
		UID       string `json:"uid"`
		Token     string `json:"token"`
		UserToken string `json:"user_token"`
	} `json:"data"`
}

type loginResp struct {
	UserID                    string             `json:"user_id"`
	UID                       string             `json:"uid"`
	Token                     string             `json:"token"`
	UserToken                 string             `json:"user_token"`
	RoleID                    string             `json:"role_id"`
	RegisterTime              int64              `json:"register_time"`
	Phone                     string             `json:"phone"`
	Email                     string             `json:"email"`
	ShowWeight                string             `json:"show_weight"`
	HasPassword               int                `json:"has_password"`
	HasDevice                 string             `json:"has_device"`
	IsOldUser                 bool               `json:"is_old_user"`
	WeightUnit                int                `json:"weightUnit"`
	HeightUnit                int                `json:"heightUnit"`
	MessageNoticeOnOff        loginMessageNotice `json:"message_notice_on_off"`
	NoLatinTurnHaveTime       string             `json:"no_latin_turn_have_time"`
	WeiboID                   string             `json:"weibo_id"`
	QQID                      string             `json:"qq_id"`
	DayimaID                  string             `json:"dayima_id"`
	BaiduID                   string             `json:"baidu_id"`
	LeyuID                    string             `json:"leyu_id"`
	JingdongID                string             `json:"jingdong_id"`
	WechatID                  string             `json:"wechat_id"`
	FacebookID                string             `json:"facebook_id"`
	FacebookName              string             `json:"facebook_name"`
	WeiboToken                string             `json:"weibo_token"`
	QQToken                   string             `json:"qq_token"`
	DayimaToken               string             `json:"dayima_token"`
	BaiduToken                string             `json:"baidu_token"`
	LeyuToken                 string             `json:"leyu_token"`
	JingdongToken             string             `json:"jingdong_token"`
	WechatToken               string             `json:"wechat_token"`
	FacebookToken             string             `json:"facebook_token"`
	AppleID                   string             `json:"apple_id"`
	AppleToken                string             `json:"apple_token"`
	AppleName                 string             `json:"apple_name"`
	HuaweiID                  string             `json:"huawei_id"`
	HuaweiToken               string             `json:"huawei_token"`
	HuaweiName                *string            `json:"huawei_name"`
	VIPType                   int                `json:"vipType"`
	StartTime                 int64              `json:"startTime"`
	EndTime                   int64              `json:"endTime"`
	Roles                     []loginRole        `json:"roles"`
	EndBodyIndex              loginBodyIndex     `json:"end_body_index"`
	PedometerData             json.RawMessage    `json:"pedometer_data"`
	PedometerLatestTime       string             `json:"pedometer_latest_time"`
	PedometerStatus           string             `json:"pedometer_status"`
	PedometerStatusUpdateTime string             `json:"pedometer_status_update_time"`
	SportLastEndTime          string             `json:"sport_last_end_time"`
	PedometerLatestDayTime    string             `json:"pedometer_latest_day_time"`
	UserDeviceList            []loginUserDevice  `json:"user_device_list"`
	LastBodyMeasure           json.RawMessage    `json:"last_body_measure"`
	TodaySportData            []json.RawMessage  `json:"todaySportData"`
	OtherBodyIndexList        []json.RawMessage  `json:"otherBodyIndexList"`
	OtherBodyMeasureList      []json.RawMessage  `json:"otherBodyMeasureList"`
	HealthkitFlag             map[string]any     `json:"healthkitFlag"`
	OtherBabyInfo             map[string]any     `json:"otherBabyInfo"`
	IsRegister                any                `json:"isRegister"`
	SMSIsEnable               int                `json:"smsIsEnable"`
	DeviceList                []loginDevice      `json:"deviceList"`
	CoachType                 int                `json:"coachType"`
	CoachLevelID              int                `json:"coachLevelId"`
	FatAlgorithmType          int                `json:"fatAlgorithmType"`
	UserType                  int                `json:"userType"`
}

type loginMessageNotice struct {
	DataMatchPush struct {
		UnknownUser int `json:"unknow_user"`
		UsualUser   int `json:"usual_user"`
		MainUser    int `json:"main_user"`
	} `json:"data_macth_push"`
	MailNoticeOnOff struct {
		SystemOnOff        int `json:"systemOnOff"`
		AppUpdateOnOff     int `json:"appUpdateOnOff"`
		AdvertisementOnOff int `json:"adviertisementOnOff"`
	} `json:"mail_notice_on_off"`
	SubscribePushOnOff struct {
		OnOff int `json:"onOff"`
	} `json:"subscribe_push_on_off"`
}

type loginRole struct {
	Type                 int               `json:"type"`
	RoleID               string            `json:"role_id"`
	RoleName             string            `json:"role_name"`
	IsAthlete            string            `json:"is_athlete"`
	Height               string            `json:"height"`
	Sex                  string            `json:"sex"`
	Birthday             string            `json:"birthday"`
	ServerTime           string            `json:"server_time"`
	UserID               int64             `json:"user_id"`
	HeadProtailURL       string            `json:"head_protail_url"`
	HeadPortraitURL      string            `json:"head_portrait_url"`
	GoalWeight           string            `json:"goal_weight"`
	GoalFat              string            `json:"goal_fat"`
	LocalTime            string            `json:"local_time"`
	FirstWeight          string            `json:"first_weight"`
	FirstFat             string            `json:"first_fat"`
	FirstUseTime         string            `json:"first_use_time"`
	ChangeGoalWeightTime string            `json:"change_goal_weight_time"`
	WeightChangeTarget   string            `json:"weight_change_target"`
	StepStateJob         string            `json:"step_state_job"`
	StepStateLife        string            `json:"step_state_life"`
	GoalStep             string            `json:"goal_step"`
	Race                 int               `json:"race"`
	AliasName            string            `json:"alias_name"`
	Email                string            `json:"email"`
	PhoneNo              string            `json:"phone_no"`
	LastBodyIndexTime    int64             `json:"last_bodyindex_time"`
	MaximumWeighingTime  int               `json:"maximum_weighing_time"`
	RoleInfos            []json.RawMessage `json:"role_infos"`
	UpgradeStatus        int               `json:"upgrade_status"`
	HeightUnit           any               `json:"heightUnit"`
	AnchorWeight         float32           `json:"anchor_weight"`
	AnchorBata           float32           `json:"anchor_bata"`
	VirtualRole          int               `json:"virtualRole"`
	BabyWeightUnit       int               `json:"babyWeightUnit"`
	RealBirthday         string            `json:"realBirthday"`
	Synopsis             any               `json:"synopsis"`
	RemarksName          any               `json:"remarksName"`
	Career               string            `json:"career"`
	Area                 string            `json:"area"`
	WeightPeriod         int               `json:"weightPeriod"`
	Cat                  int               `json:"cat"`
	UserType             any               `json:"userType"`
}

type loginBodyIndex struct {
	BodyTime                 int64         `json:"bodyTime"`
	DataType                 int           `json:"dataType"`
	BodyIndexID              int64         `json:"body_index_id"`
	RoleID                   int64         `json:"role_id"`
	BodyFat                  float32       `json:"body_fat"`
	Weight                   float32       `json:"weight"`
	BMI                      float32       `json:"bmi"`
	VisceralFatLevel         float32       `json:"visceral_fat_level"`
	MuscleRace               float32       `json:"muscle_race"`
	BodyAge                  float32       `json:"body_age"`
	BoneMass                 float32       `json:"bone_mass"`
	BasicMetabolism          float32       `json:"basic_metabolism"`
	WaterRace                float32       `json:"water_race"`
	SkeletalMuscle           float32       `json:"skeletal_muscle"`
	LocalTime                int64         `json:"local_time"`
	SubcutaneousFat          float32       `json:"subcutaneous_fat"`
	ServerTime               int64         `json:"server_time"`
	ServerID                 int64         `json:"server_id"`
	IsDel                    int           `json:"is_del"`
	Abnormal                 loginAbnormal `json:"abnormal"`
	AbnormalFlag             int           `json:"abnormal_flag"`
	ElectricResistance       float32       `json:"electric_resistance"`
	IsManuallyAdd            int           `json:"is_manually_add"`
	IsFirstDay               int           `json:"is_first_day"`
	LandmarkIcons            []string      `json:"landmarkIcons"`
	LandmarkIconsV2          []string      `json:"landmarkIconsV2"`
	MAC                      string        `json:"mac"`
	AnchorWeight             float32       `json:"anchor_weight"`
	AnchorBata               float32       `json:"anchor_bata"`
	CorrectionValueR         float32       `json:"correction_value_r"`
	BodyFatReferenceValue    float32       `json:"body_fat_reference_value"`
	LabelMarker              int           `json:"label_marker"`
	DataSources              int           `json:"data_sources"`
	ElectricResistanceFilter float32       `json:"electric_resistance_filter"`
	BodyFatOriginal          float32       `json:"body_fat_original"`
	Noise                    float32       `json:"noise"`
	FatAlgorithmType         int           `json:"fat_algorithm_type"`
	Cat                      int           `json:"cat"`
	VerifiedMark             int           `json:"verifiedMark"`
	FirmwareVersion          string        `json:"firmwareVersion"`
}

type loginAbnormal struct {
	Weight       float32 `json:"weight"`
	Time         int64   `json:"time"`
	AbnormalFlag int     `json:"abnormal_flag"`
	BodyFat      float32 `json:"body_fat"`
}

type loginUserDevice struct {
	UserID         int64  `json:"user_id"`
	LatinName      string `json:"latin_name"`
	LatinModel     int    `json:"latin_model"`
	LatinMAC       string `json:"latin_mac"`
	ShowWeight     int    `json:"show_weight"`
	BindClientTime int64  `json:"bind_client_time"`
	BindServerTime int64  `json:"bind_server_time"`
}

type loginDevice struct {
	DeviceID          int    `json:"deviceId"`
	Name              string `json:"name"`
	UserDeviceName    string `json:"userDeviceName"`
	BroadcastName     string `json:"broadcastName"`
	AttendMode        int    `json:"attendMode"`
	WeightUnit        []int  `json:"weightUnit"`
	UserManagement    int    `json:"userManagement"`
	WifiSet           int    `json:"wifiSet"`
	OTA               int    `json:"ota"`
	Privacy           int    `json:"privacy"`
	UnitSwitch        int    `json:"unitSwitch"`
	PowerMeasurement  int    `json:"powerMeasurement"`
	FrontViewURL      string `json:"frontViewUrl"`
	FrontyFiveViewURL string `json:"frontyFiveViewUrl"`
	MatchBalanceURL   string `json:"matchBalanceUrl"`
	Order             int    `json:"order"`
	Area              int    `json:"area"`
	PrivacyOnOff      int    `json:"privacyOnOff"`
	BindClientTime    int64  `json:"bindClientTime"`
	BindServerTime    int64  `json:"bindServerTime"`
	UserID            int64  `json:"userId"`
	MAC               string `json:"mac"`
	LowScale          bool   `json:"lowScale"`
	Brand             string `json:"brand"`
	Balance           int    `json:"balance"`
	PhaseAngle        int    `json:"phaseAngle"`
}

func (c *Client) Login(username, password string) error {
	form := c.values("user_login_new")

	var req1 struct {
		AppVer    string `json:"appver"`
		Timestamp string `json:"timestamp"`
		Lang      string `json:"lang"`
		Method    string `json:"method"`
		Timezone  string `json:"timezone"`
		Sign      string `json:"sign"`
		PushToken string `json:"push_token"`
		DeviceID  string `json:"device_id"`
		Rec       struct {
			AppChannel  string `json:"app_channel"`
			AppVer      string `json:"app_version"`
			Email       string `json:"email"`
			Password    string `json:"password"`
			Phone       string `json:"phone"`
			PhoneSystem string `json:"phone_system"`
			PhoneType   string `json:"phone_type"`
		} `json:"req"`
	}
	req1.AppVer = form.Get("appver")
	req1.Timestamp = form.Get("timestamp")
	req1.Lang = form.Get("lang")
	req1.Method = form.Get("method")
	req1.Sign = form.Get("sign")
	req1.PushToken = form.Get("push_token")
	req1.DeviceID = form.Get("device_id")
	req1.Rec.AppVer = form.Get("appver")
	req1.Rec.Email = username
	req1.Rec.Password = password

	data, err := json.Marshal(req1)
	if err != nil {
		return err
	}

	form.Set("reqData", string(data))

	res, err := c.client.Post(
		api+"v1/api/account/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var res1 loginResponse

	if err = json.Unmarshal(body, &res1); err != nil {
		return err
	}

	if res1.Code != 0 {
		return errors.New("picooc: login error: " + res1.Msg)
	}

	c.userID = firstNonEmpty(res1.Resp.UserID, res1.Resp.UID, res1.Data.UserID, res1.Data.UID, res1.UserID, res1.UID)
	c.token = firstNonEmpty(res1.Resp.Token, res1.Resp.UserToken, res1.Data.Token, res1.Data.UserToken, res1.Token, res1.UserToken)

	c.roleIDs = map[string]string{}
	c.roles = nil
	for _, role := range res1.Resp.Roles {
		if role.RoleID == "" {
			continue
		}
		c.roles = append(c.roles, roleInfo{ID: role.RoleID, Name: role.RoleName})
		c.roleIDs[role.RoleName] = role.RoleID
		log.Printf("picooc: found roleid=%s rolename=%s\n", role.RoleID, role.RoleName)
	}

	if len(c.roles) == 0 && res1.Resp.RoleID != "" {
		c.roles = append(c.roles, roleInfo{ID: res1.Resp.RoleID, Name: ""})
		log.Printf("picooc: found roleid=%s rolename=%s\n", res1.Resp.RoleID, "")
	}

	return nil
}

const appVer = "i4.1.11.0"

func (c *Client) values(method string) url.Values {
	return c.valuesWithAppVer(method, appVer)
}

func (c *Client) valuesWithAppVer(method, version string) url.Values {
	if c.deviceID == "" {
		c.deviceID = strings.ToUpper(uuid.NewString())
	}

	timestamp := strconv.Itoa(int(time.Now().Unix()))
	sign := upperMD5(c.deviceID + upperMD5(timestamp+upperMD5(method)+upperMD5(version)))

	return url.Values{
		"appver":     {version},
		"timestamp":  {timestamp},
		"lang":       {"en"},
		"method":     {method},
		"timezone":   {""}, // don't know how to get right value
		"sign":       {sign},
		"push_token": {"android::" + c.deviceID},
		"device_id":  {c.deviceID},
	}
}

func upperMD5(s string) string {
	return fmt.Sprintf("%X", md5.Sum([]byte(s)))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
