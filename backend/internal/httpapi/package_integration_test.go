package httpapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
)

func TestPackageDraftPublicationAndAgencyIsolation(t *testing.T) {
	db := testutil.Database(t)
	cfg := testConfig()
	cfg.AuthRatePerMinute, cfg.APIRatePerMinute = 1000, 10000
	handler, err := New(db, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	hash, _ := security.HashPassword("package-tests-password", 10)
	accounts := []models.User{
		{Email: "agency-one@packages.test", Role: models.RoleAgency},
		{Email: "agency-two@packages.test", Role: models.RoleAgency},
		{Email: "tourist@packages.test", Role: models.RoleUser},
		{Email: "admin@packages.test", Role: models.RoleAdmin},
		{Email: "attraction@packages.test", Role: models.RoleAttraction},
	}
	for i := range accounts {
		accounts[i].FirstName, accounts[i].LastName, accounts[i].PasswordHash, accounts[i].Status = "Prueba", "Paquete", hash, models.StatusActive
		if err := db.Create(&accounts[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	agencies := []models.Agency{
		{Name: "Agencia Uno", Department: "Tarija", City: "Tarija", Address: "Centro", Phone: "70000001", Email: "uno@agency.test", ManagerID: accounts[0].ID, Status: models.StatusActive, Published: true, MinimumPayingAge: 6, Version: 1},
		{Name: "Agencia Dos", Department: "Tarija", City: "Tarija", Address: "Centro", Phone: "70000002", Email: "dos@agency.test", ManagerID: accounts[1].ID, Status: models.StatusActive, MinimumPayingAge: 8, Version: 1},
	}
	for i := range agencies {
		if err := db.Create(&agencies[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	attraction := models.Attraction{Name: "Valle", Description: "Atracción pública", Category: "Natural", Department: "Tarija", City: "Tarija", Address: "Camino", ManagerID: accounts[4].ID, Status: models.StatusActive, Published: true, Version: 1, ScheduleMode: "unspecified", SeasonMode: "unspecified", OpeningDays: []int{}}
	if err := db.Create(&attraction).Error; err != nil {
		t.Fatal(err)
	}

	cookies := make([]*http.Cookie, len(accounts))
	csrf := make([]string, len(accounts))
	rawRequest := func(method, path string, body any, account int) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Origin", "http://localhost:3000")
		if account >= 0 && cookies[account] != nil {
			req.AddCookie(cookies[account])
			req.Header.Set("X-CSRF-Token", csrf[account])
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	request := func(method, path string, body any, account, want int) *httptest.ResponseRecorder {
		t.Helper()
		response := rawRequest(method, path, body, account)
		if response.Code != want {
			t.Fatalf("%s %s got %d want %d: %s", method, path, response.Code, want, response.Body.String())
		}
		return response
	}
	for i, account := range accounts {
		response := request("POST", "/api/v1/auth/login", map[string]string{"email": account.Email, "password": "package-tests-password"}, -1, 200)
		cookies[i] = response.Result().Cookies()[0]
		var envelope struct {
			Data struct {
				CSRF string `json:"csrf_token"`
			}
		}
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		csrf[i] = envelope.Data.CSRF
	}
	read := func(response *httptest.ResponseRecorder) models.TourPackage {
		t.Helper()
		var envelope struct{ Data models.TourPackage }
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return envelope.Data
	}
	readDeparture := func(response *httptest.ResponseRecorder) models.TourPackageDeparture {
		t.Helper()
		var envelope struct{ Data models.TourPackageDeparture }
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return envelope.Data
	}
	draftPayload := map[string]any{"name": "Ruta chapaca", "duration_days": 2, "duration_nights": 1}
	request("POST", "/api/v1/agency/packages", draftPayload, 2, 403)
	request("GET", "/api/v1/agency/packages", nil, 3, 403)
	request("GET", "/api/v1/agency/packages", nil, 4, 403)
	draft := read(request("POST", "/api/v1/agency/packages", draftPayload, 0, 201))
	if draft.Published || draft.Version != 1 || draft.AgencyID != agencies[0].ID {
		t.Fatalf("wrong draft: %+v", draft)
	}
	path := fmt.Sprintf("/api/v1/agency/packages/%d", draft.ID)
	publicPath := fmt.Sprintf("/api/v1/packages/%d", draft.ID)
	request("GET", publicPath, nil, -1, 404)
	request("GET", "/api/v1/packages?difficulty=impossible", nil, -1, 422)
	request("GET", "/api/v1/packages?max_price_cents=-1", nil, -1, 422)
	request("GET", "/api/v1/packages?date_from=not-a-date", nil, -1, 422)
	hidden := models.TourPackage{AgencyID: agencies[1].ID, Name: "Paquete oculto", Description: "No debe aparecer porque la agencia no está publicada.", DurationDays: 1, NationalPriceCents: 10000, Published: true, Version: 1, Includes: []string{}, Excludes: []string{}, Bring: []string{}}
	if err := db.Create(&hidden).Error; err != nil {
		t.Fatal(err)
	}
	request("GET", fmt.Sprintf("/api/v1/packages/%d", hidden.ID), nil, -1, 404)
	request("GET", path, nil, 1, 404)
	request("PUT", path, map[string]any{"name": "Ruta chapaca", "duration_days": 2, "duration_nights": 1, "published": true, "version": 1}, 0, 422)

	var pngData bytes.Buffer
	if err := png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	futureDate := time.Now().In(time.FixedZone("America/La_Paz", -4*60*60)).AddDate(0, 0, 7).Format("2006-01-02")
	publishedPayload := map[string]any{
		"name": "Ruta chapaca", "description": "Dos días para descubrir paisajes y cultura de Tarija.", "duration_days": 2, "duration_nights": 1,
		"national_price_cents": 25000, "foreign_surcharge_cents": 5000, "cancellation_allowed": true, "cancellation_notice_hours": 48, "published": true, "version": 1,
		"photos": []map[string]any{{"data": "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData.Bytes())}},
		"itinerary": []map[string]any{
			{"day_number": 1, "title": "Bienvenida", "attraction_ids": []uint64{attraction.ID}},
			{"day_number": 2, "title": "Despedida", "attraction_ids": []uint64{attraction.ID}},
		},
		"schedule": map[string]any{
			"frequency_type": "single", "valid_from": futureDate, "valid_until": futureDate,
			"departure_time": "08:30", "meeting_time": "08:00", "default_min_capacity": 4, "default_max_capacity": 12,
			"booking_cutoff_hours": 24, "maximum_advance_days": 180, "default_meeting_point": "Plaza principal",
		},
	}
	published := read(request("PUT", path, publishedPayload, 0, 200))
	if !published.Published || published.Version != 2 || len(published.Photos) != 1 || len(published.Itinerary) != 2 || published.Schedule == nil || len(published.Departures) != 1 {
		t.Fatalf("publication failed: %+v", published)
	}
	var catalogEnvelope struct {
		Data struct {
			Packages []struct {
				ID            uint64 `json:"id"`
				AgencyName    string `json:"agency_name"`
				Bookable      bool   `json:"bookable"`
				NextDeparture *struct {
					ID                uint64 `json:"id"`
					AvailableCapacity int    `json:"available_capacity"`
					Bookable          bool   `json:"bookable"`
				} `json:"next_departure"`
			} `json:"packages"`
		} `json:"data"`
	}
	catalogResponse := request("GET", "/api/v1/packages?department=Tarija&bookable=true", nil, -1, 200)
	if err := json.Unmarshal(catalogResponse.Body.Bytes(), &catalogEnvelope); err != nil {
		t.Fatal(err)
	}
	if len(catalogEnvelope.Data.Packages) != 1 || catalogEnvelope.Data.Packages[0].ID != published.ID || catalogEnvelope.Data.Packages[0].AgencyName != agencies[0].Name || !catalogEnvelope.Data.Packages[0].Bookable || catalogEnvelope.Data.Packages[0].NextDeparture == nil || !catalogEnvelope.Data.Packages[0].NextDeparture.Bookable || catalogEnvelope.Data.Packages[0].NextDeparture.AvailableCapacity != 12 {
		t.Fatalf("wrong public catalog: %s", catalogResponse.Body.String())
	}
	var publicDetailEnvelope struct {
		Data struct {
			ID               uint64 `json:"id"`
			MinimumPayingAge int    `json:"minimum_paying_age"`
			Itinerary        []struct {
				Attractions []any `json:"attractions"`
			} `json:"itinerary"`
			Departures []struct {
				ID                  uint64 `json:"id"`
				RemainingForMinimum int    `json:"remaining_for_minimum"`
			} `json:"departures"`
		} `json:"data"`
	}
	publicDetailResponse := request("GET", publicPath, nil, -1, 200)
	if err := json.Unmarshal(publicDetailResponse.Body.Bytes(), &publicDetailEnvelope); err != nil {
		t.Fatal(err)
	}
	if publicDetailEnvelope.Data.ID != published.ID || publicDetailEnvelope.Data.MinimumPayingAge != 6 || len(publicDetailEnvelope.Data.Itinerary) != 2 || len(publicDetailEnvelope.Data.Departures) != 1 || publicDetailEnvelope.Data.Departures[0].RemainingForMinimum != 4 {
		t.Fatalf("wrong public detail: %s", publicDetailResponse.Body.String())
	}
	request("GET", fmt.Sprintf("%s/photos/%d", publicPath, published.Photos[0].ID), nil, -1, 200)
	departure := published.Departures[0]
	departurePath := fmt.Sprintf("%s/departures/%d", path, departure.ID)
	adjustment := map[string]any{
		"version": departure.Version, "meeting_time": "07:45", "min_capacity": 5, "max_capacity": 15,
		"booking_cutoff_hours": 48, "meeting_point": "Terminal de buses", "instructions": "Presentarse con documento.",
	}
	request("PATCH", departurePath, adjustment, 1, 404)
	start := make(chan struct{})
	responses := make(chan *httptest.ResponseRecorder, 2)
	for range 2 {
		go func() {
			<-start
			responses <- rawRequest("PATCH", departurePath, adjustment, 0)
		}()
	}
	close(start)
	var success *httptest.ResponseRecorder
	statusCounts := map[int]int{}
	for range 2 {
		response := <-responses
		statusCounts[response.Code]++
		if response.Code == http.StatusOK {
			success = response
		}
	}
	if statusCounts[http.StatusOK] != 1 || statusCounts[http.StatusConflict] != 1 || success == nil {
		t.Fatalf("concurrent departure update got statuses: %v", statusCounts)
	}
	adjusted := readDeparture(success)
	if !adjusted.IsException || adjusted.Version != departure.Version+1 || adjusted.MinCapacity != 5 || adjusted.MaxCapacity != 15 || adjusted.MeetingPoint != "Terminal de buses" || adjusted.ModifiedByUserID == nil {
		t.Fatalf("departure adjustment failed: %+v", adjusted)
	}
	publishedPayload["version"] = published.Version
	publishedPayload["photos"] = []map[string]any{{"id": published.Photos[0].ID}}
	regenerated := read(request("PUT", path, publishedPayload, 0, 200))
	if len(regenerated.Departures) != 1 || !regenerated.Departures[0].IsException || regenerated.Departures[0].MaxCapacity != 15 || regenerated.Departures[0].MeetingPoint != "Terminal de buses" {
		t.Fatalf("schedule regeneration overwrote departure exception: %+v", regenerated.Departures)
	}
	adjusted = regenerated.Departures[0]
	request("POST", departurePath+"/cancel", map[string]any{"version": adjusted.Version, "reason": "No"}, 0, 422)
	cancelled := readDeparture(request("POST", departurePath+"/cancel", map[string]any{"version": adjusted.Version, "reason": "Condiciones climáticas adversas"}, 0, 200))
	if cancelled.Status != "cancelled" || cancelled.CancellationReason == "" || cancelled.CancelledAt == nil || !cancelled.IsException || cancelled.Version != adjusted.Version+1 {
		t.Fatalf("departure cancellation failed: %+v", cancelled)
	}
	publicDetailResponse = request("GET", publicPath, nil, -1, 200)
	if err := json.Unmarshal(publicDetailResponse.Body.Bytes(), &publicDetailEnvelope); err != nil {
		t.Fatal(err)
	}
	if len(publicDetailEnvelope.Data.Departures) != 0 {
		t.Fatalf("cancelled departure remained public: %s", publicDetailResponse.Body.String())
	}
	request("PATCH", departurePath, adjustment, 0, 409)
	photoPath := fmt.Sprintf("%s/photos/%d", path, published.Photos[0].ID)
	request("GET", photoPath, nil, 1, 404)
	request("PUT", path, publishedPayload, 0, 409)
	if err := db.Model(&models.Attraction{}).Where("id = ?", attraction.ID).Update("published", false).Error; err != nil {
		t.Fatal(err)
	}
	publicDetailResponse = request("GET", publicPath, nil, -1, 200)
	if err := json.Unmarshal(publicDetailResponse.Body.Bytes(), &publicDetailEnvelope); err != nil {
		t.Fatal(err)
	}
	for _, day := range publicDetailEnvelope.Data.Itinerary {
		if len(day.Attractions) != 0 {
			t.Fatalf("unpublished attraction remained linked in public itinerary: %s", publicDetailResponse.Body.String())
		}
	}
	if err := db.Model(&models.Agency{}).Where("id = ?", agencies[0].ID).Update("status", models.StatusInactive).Error; err != nil {
		t.Fatal(err)
	}
	request("GET", publicPath, nil, -1, 404)
	request("GET", fmt.Sprintf("%s/photos/%d", publicPath, published.Photos[0].ID), nil, -1, 404)
}
