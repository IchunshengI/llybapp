package bazi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "llyb-backend/proto"
)

// Reasoning is the backend handler for the "基础推理" page.
//
// For now it just echoes the request payload back to the caller so the frontend
// can validate wiring without depending on real bazi/solar-time logic yet.
func Reasoning(ctx context.Context, req *pb.ReasoningRequest) (*pb.ReasoningResponse, error) {
	if req == nil {
		return &pb.ReasoningResponse{
			Code:    1002,
			Message: "参数不合法",
		}, nil
	}

	trueSolarTime := ""
	province := req.GetProvince()
	city := req.GetCity()

	lonDeg, lonErr := ResolveCityLongitude(ctx, province, city)
	lonOK := lonErr == nil
	lonSource := "amap"
	var trueSolarTimeErr string
	if lonOK {
		if s, err := TrueSolarTimeFromBeijing(req.GetSolarDate(), req.GetBirthTime(), lonDeg); err == nil {
			trueSolarTime = s
		} else {
			trueSolarTimeErr = err.Error()
		}
	} else {
		// If AMap fails, fall back to Beijing time and surface the AMap error.
		bt, err := time.ParseInLocation(
			"2006-01-02 15:04",
			req.GetSolarDate()+" "+req.GetBirthTime(),
			time.FixedZone("CST", 8*3600),
		)
		if err == nil {
			trueSolarTime = bt.Format("2006/01/02 15:04")
		}
		trueSolarTimeErr = "amap_failed: " + lonErr.Error()
	}

	echo := map[string]any{
		"gender":     req.GetGender().String(),
		"solar_date": req.GetSolarDate(),
		"birth_time": req.GetBirthTime(),
		"province":   province,
		"city":       city,
		"longitude_deg": func() any {
			if lonOK {
				return lonDeg
			}
			return nil
		}(),
		"longitude_source": func() any {
			if lonOK {
				return lonSource
			}
			return nil
		}(),
		"true_solar_time":     trueSolarTime, // "YYYY/MM/DD HH:mm" (empty if longitude not resolved)
		"true_solar_time_err": trueSolarTimeErr,
	}

	b, err := json.Marshal(echo)
	if err != nil {
		return &pb.ReasoningResponse{
			Code:    1003,
			Message: fmt.Sprintf("json marshal failed: %v", err),
		}, nil
	}

	return &pb.ReasoningResponse{
		Code:       0,
		Message:    "ok",
		ResultJson: string(b),
	}, nil
}
