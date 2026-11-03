package driver

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/kesiai/sdk-go/conn/mq"
	"github.com/kesiai/sdk-go/driver/entity"
)

type recordingMQ struct {
	mq.MQ
	topic   []string
	payload []byte
}

func (m *recordingMQ) Publish(ctx context.Context, topic []string, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.topic = append([]string(nil), topic...)
	m.payload = append([]byte(nil), payload...)
	return nil
}

func testApp(t *testing.T) (*app, *recordingMQ) {
	t.Helper()
	previous := Cfg
	Cfg = new(Config)
	Cfg.Project = "test-project"
	Cfg.MQ.Timeout = time.Second
	t.Cleanup(func() { Cfg = previous })
	broker := new(recordingMQ)
	return &app{mq: broker}, broker
}

func assertPublishedPoint(t *testing.T, broker *recordingMQ) entity.WritePoint {
	t.Helper()
	if want := []string{"data", "test-project", "test", "b1"}; !reflect.DeepEqual(broker.topic, want) {
		t.Fatalf("topic = %v, want %v", broker.topic, want)
	}
	var point entity.WritePoint
	if err := json.Unmarshal(broker.payload, &point); err != nil {
		t.Fatal(err)
	}
	if point.ID != "b1" || point.Source != "device" || point.UnixTime == 0 {
		t.Fatalf("unexpected published point: %+v", point)
	}
	if _, ok := point.Fields["p1"]; !ok {
		t.Fatalf("missing p1: %+v", point.Fields)
	}
	return point
}

func TestApp_WritePoints(t *testing.T) {
	a, broker := testApp(t)
	var minValue float64 = 10
	var MaxValue float64 = 100
	//var MinRaw float64 = 200
	//var MaxRaw float64 = 300
	var Fixed int32 = 2
	var Mod float64 = 2
	point := entity.Point{
		Table: "test",
		ID:    "b1",
		Fields: []entity.Field{
			{Tag: entity.Tag{
				ID:       "p1",
				Name:     "p1",
				TagValue: &entity.TagValue{
					//MinValue: &minValue,
					//MaxValue: &MaxValue,
					//MinRaw:   &MinRaw,
					//MaxRaw:   &MaxRaw,
				},
				Fixed: &Fixed,
				Mod:   &Mod,
				Range: &entity.Range{
					MinValue:   &minValue,
					MaxValue:   &MaxValue,
					Active:     entity.Active_Boundary,
					FixedValue: &Mod,
				},
			},
				Value: 10.121,
			},
		},
	}

	err := a.WritePoints(context.Background(), point)
	if err != nil {
		t.Fatal(err)
	}
	assertPublishedPoint(t, broker)
}

func Test_wp(t *testing.T) {
	a, broker := testApp(t)
	var minValue float64 = 10
	var MaxValue float64 = 100
	//var MinRaw float64 = 200
	//var MaxRaw float64 = 300
	var Fixed int32 = 2
	var Mod float64 = 2
	point := entity.Point{
		ID: "b1",
		Fields: []entity.Field{
			{Tag: entity.Tag{
				ID:       "p1",
				Name:     "p1",
				TagValue: &entity.TagValue{
					//MinValue: &minValue,
					//MaxValue: &MaxValue,
					//MinRaw:   &MinRaw,
					//MaxRaw:   &MaxRaw,
				},
				Fixed: &Fixed,
				Mod:   &Mod,
				Range: &entity.Range{
					MinValue:   &minValue,
					MaxValue:   &MaxValue,
					Active:     entity.Active_Boundary,
					FixedValue: &Mod,
				},
			},
				Value: true,
			},
		},
	}
	err := a.writePoints(context.Background(), "test", point)
	if err != nil {
		t.Fatal(err)
	}
	got := assertPublishedPoint(t, broker)
	if got.Fields["p1"] != float64(1) {
		t.Fatalf("p1 = %v, want 1", got.Fields["p1"])
	}
}
