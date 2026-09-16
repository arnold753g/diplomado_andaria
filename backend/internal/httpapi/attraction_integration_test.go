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
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"starter-backend/internal/models"
	"starter-backend/internal/security"
	"starter-backend/internal/testutil"
)

func TestAttractionCatalogOwnershipAndFavorites(t *testing.T) {
	db := testutil.Database(t)
	cfg := testConfig()
	cfg.AuthRatePerMinute = 1000
	cfg.APIRatePerMinute = 10000
	handler, err := New(db, cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	hash, err := security.HashPassword("agency-tests-password", 10)
	if err != nil {
		t.Fatal(err)
	}
	accounts := []models.User{
		{Email: "admin@agency.test", Role: models.RoleAdmin},
		{Email: "manager1@agency.test", Role: models.RoleAttraction},
		{Email: "manager2@agency.test", Role: models.RoleAttraction},
		{Email: "tourist@agency.test", Role: models.RoleUser},
		{Email: "attraction@agency.test", Role: models.RoleAgency},
		{Email: "manager3@agency.test", Role: models.RoleAttraction},
	}
	for i := range accounts {
		accounts[i].FirstName = "Prueba"
		accounts[i].LastName = "Agencia"
		accounts[i].PasswordHash = hash
		accounts[i].Status = models.StatusActive
		if err := db.Create(&accounts[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	cookies := make([]*http.Cookie, len(accounts))
	csrf := make([]string, len(accounts))
	rawRequest := func(method, path string, body any, account int, withCSRF bool) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		r := httptest.NewRequest(method, path, bytes.NewReader(raw))
		r.Header.Set("Origin", "http://localhost:3000")
		if account >= 0 && cookies[account] != nil {
			r.AddCookie(cookies[account])
			if withCSRF {
				r.Header.Set("X-CSRF-Token", csrf[account])
			}
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	request := func(method, path string, body any, account, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := rawRequest(method, path, body, account, true)
		if w.Code != want {
			t.Fatalf("%s %s got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		return w
	}
	for i, account := range accounts {
		w := request("POST", "/api/v1/auth/login", map[string]string{"email": account.Email, "password": "agency-tests-password"}, -1, 200)
		cookies[i] = w.Result().Cookies()[0]
		var response struct {
			Data struct {
				CSRF string `json:"csrf_token"`
			}
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		csrf[i] = response.Data.CSRF
	}

	type view struct {
		models.Attraction
		ManagerID uint64 `json:"manager_id"`
		Status    string `json:"status"`
		Published bool   `json:"published"`
		Version   int    `json:"version"`
	}
	read := func(w *httptest.ResponseRecorder) view {
		t.Helper()
		var result struct{ Data view }
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result.Data
	}
	total := func(w *httptest.ResponseRecorder) int {
		t.Helper()
		var result struct {
			Data struct{ Pagination struct{ Total int } }
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result.Data.Pagination.Total
	}
	var pngBytes bytes.Buffer
	if err := png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 12, 12))); err != nil {
		t.Fatal(err)
	}
	photoData := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes.Bytes())
	payload := map[string]any{"name": "Valle de prueba", "description": "Un lugar para conocer la naturaleza de Tarija.", "subcategory_ids": []uint64{14, 1}, "department": "Tarija", "city": "Tarija", "address": "Camino de prueba 123", "manager_id": accounts[1].ID}
	request("GET", "/api/v1/attractions/options", nil, -1, 200)
	request("POST", "/api/v1/admin/attractions", payload, -1, 401)
	request("POST", "/api/v1/admin/attractions", payload, 1, 403)
	if w := rawRequest("POST", "/api/v1/admin/attractions", payload, 0, false); w.Code != 403 {
		t.Fatal("missing CSRF accepted")
	}
	invalid := maps.Clone(payload)
	invalid["manager_id"] = accounts[4].ID
	request("POST", "/api/v1/admin/attractions", invalid, 0, 422)
	first := read(request("POST", "/api/v1/admin/attractions", payload, 0, 201))
	if first.Published || first.Version != 1 {
		t.Fatal("new attraction must be a draft")
	}
	// Multiple attractions may share exactly one manager.
	second := read(request("POST", "/api/v1/admin/attractions", payload, 0, 201))
	if total(request("GET", "/api/v1/managed-attractions", nil, 1, 200)) != 2 {
		t.Fatal("multiple assignments not visible")
	}
	if total(request("GET", "/api/v1/managed-attractions", nil, 2, 200)) != 0 {
		t.Fatal("another manager can list private attractions")
	}
	private := fmt.Sprintf("/api/v1/managed-attractions/%d", first.ID)
	public := fmt.Sprintf("/api/v1/attractions/%d", first.ID)
	adminPath := fmt.Sprintf("/api/v1/admin/attractions/%d", first.ID)
	request("GET", private, nil, 2, 404)
	request("GET", private, nil, 3, 403)
	request("GET", private, nil, 4, 403)
	request("GET", public, nil, -1, 404)
	if total(request("GET", "/api/v1/attractions", nil, -1, 200)) != 0 {
		t.Fatal("draft leaked into catalog")
	}
	own := maps.Clone(payload)
	delete(own, "manager_id")
	own["version"] = 1
	own["published"] = true
	own["schedule_mode"] = "scheduled"
	own["opening_time"] = "08:15"
	own["closing_time"] = "18:45"
	own["opening_days"] = []int{7, 1, 3}
	own["season_mode"] = "months"
	own["season_start_month"] = 11
	own["season_end_month"] = 2
	request("PUT", private, own, 1, 422) // Publishing requires a photo.
	own["photos"] = []map[string]any{{"data": photoData}}
	forged := maps.Clone(own)
	forged["manager_id"] = accounts[2].ID
	request("PUT", private, forged, 1, 403)
	forged = maps.Clone(own)
	forged["status"] = "inactive"
	request("PUT", private, forged, 1, 403)
	for _, change := range []map[string]any{{"latitude": 91, "longitude": -64}, {"latitude": -21}, {"admission_cents": -1}, {"admission_cents": 12.5}, {"subcategory_ids": []uint64{99999}}, {"photos": []map[string]any{{"data": "data:image/svg+xml;base64,PHN2Zy8+"}}}} {
		bad := maps.Clone(own)
		for k, v := range change {
			bad[k] = v
		}
		expected := 422
		if _, ok := change["admission_cents"].(float64); ok {
			expected = 400
		}
		request("PUT", private, bad, 1, expected)
	}
	published := read(request("PUT", private, own, 1, 200))
	if !published.Published || len(published.Photos) != 1 || published.Version != 2 {
		t.Fatal("publication failed")
	}
	if len(published.Subcategories) != 2 || published.Subcategories[0].SubcategoryID != 14 || published.Category != "Natural" {
		t.Fatal("classification order not preserved")
	}
	if published.OpeningTime != "08:15" || len(published.OpeningDays) != 3 || published.OpeningDays[0] != 1 || published.SeasonStartMonth == nil || *published.SeasonStartMonth != 11 || published.SeasonEndMonth == nil || *published.SeasonEndMonth != 2 {
		t.Fatal("schedule/season not persisted")
	}
	request("PUT", private, own, 1, 409)
	pub := request("GET", public, nil, -1, 200)
	for _, secret := range []string{"manager_id", "manager_email", "manager_name", "password", "version", "published"} {
		if bytes.Contains(pub.Body.Bytes(), []byte(secret)) {
			t.Fatalf("public response disclosed %s", secret)
		}
	}
	photoURL := fmt.Sprintf("%s/photos/%d", public, published.Photos[0].ID)
	img := request("GET", photoURL, nil, -1, 200)
	if img.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatal("photo was not normalized")
	}
	if _, _, err := image.Decode(bytes.NewReader(img.Body.Bytes())); err != nil {
		t.Fatal(err)
	}
	request("GET", fmt.Sprintf("/api/v1/attractions/%d/photos/%d", second.ID, published.Photos[0].ID), nil, -1, 404)
	if total(request("GET", "/api/v1/attractions?department=Tarija&category=Natural&search=valle", nil, -1, 200)) != 1 {
		t.Fatal("catalog filters failed")
	}
	if total(request("GET", "/api/v1/attractions?search=%25", nil, -1, 200)) != 0 {
		t.Fatal("search wildcard was not escaped")
	}
	if total(request("GET", "/api/v1/attractions?category=Enoturismo&subcategory_id=1", nil, -1, 200)) != 1 {
		t.Fatal("secondary classification filter failed")
	}
	if total(request("GET", "/api/v1/attractions?category=Natural&subcategory_id=2", nil, -1, 200)) != 0 {
		t.Fatal("subcategory filter ignored")
	}
	for _, change := range []map[string]any{
		{"subcategory_ids": []uint64{}}, {"subcategory_ids": []uint64{1, 1}}, {"subcategory_ids": []uint64{1, 2, 3, 4, 5}},
		{"opening_time": "25:10"}, {"closing_time": "08:15"}, {"opening_days": []int{1, 1}}, {"opening_days": []int{0}}, {"opening_days": []int{}},
		{"season_start_month": 0}, {"season_end_month": nil}, {"season_mode": "all_year"},
	} {
		bad := maps.Clone(own)
		bad["version"] = 2
		for k, v := range change {
			bad[k] = v
		}
		request("PUT", private, bad, 1, 422)
	}
	// Cannot steal another attraction's photos, and the entire failed edit rolls back.
	other := maps.Clone(own)
	other["version"] = 1
	other["photos"] = []map[string]any{{"id": published.Photos[0].ID}}
	request("PUT", fmt.Sprintf("/api/v1/managed-attractions/%d", second.ID), other, 1, 422)
	if read(request("GET", fmt.Sprintf("/api/v1/managed-attractions/%d", second.ID), nil, 1, 200)).Version != 1 {
		t.Fatal("failed photo update was not atomic")
	}
	favoritePath := fmt.Sprintf("/api/v1/me/favorites/%d", first.ID)
	request("PUT", favoritePath, nil, -1, 401)
	request("PUT", favoritePath, nil, 1, 403)
	request("PUT", favoritePath, nil, 0, 403)
	if w := rawRequest("PUT", favoritePath, nil, 3, false); w.Code != 403 {
		t.Fatal("favorite accepted without CSRF")
	}
	// Concurrent, repeated favorite requests are idempotent.
	var wg sync.WaitGroup
	results := make(chan int, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- rawRequest("PUT", favoritePath, nil, 3, true).Code }()
	}
	wg.Wait()
	close(results)
	for code := range results {
		if code != 200 {
			t.Fatalf("favorite concurrency failed: %d", code)
		}
	}
	if total(request("GET", "/api/v1/me/favorites", nil, 3, 200)) != 1 {
		t.Fatal("duplicate or missing favorites")
	}
	if !bytes.Contains(request("GET", favoritePath, nil, 3, 200).Body.Bytes(), []byte(`"favorite":true`)) {
		t.Fatal("favorite state missing")
	}
	request("DELETE", favoritePath, nil, 3, 200)
	request("DELETE", favoritePath, nil, 3, 200)
	if total(request("GET", "/api/v1/me/favorites", nil, 3, 200)) != 0 {
		t.Fatal("favorite not removed")
	}
	request("PUT", favoritePath, nil, 3, 200)
	request("PATCH", fmt.Sprintf("/api/v1/admin/users/%d/role", accounts[1].ID), map[string]string{"role": models.RoleUser}, 0, 409)
	request("PATCH", fmt.Sprintf("/api/v1/admin/users/%d", accounts[1].ID), map[string]string{"first_name": "Test", "last_name": "Manager", "role": models.RoleUser, "status": "active"}, 0, 409)
	// Competing saves: one succeeds and one receives a version conflict.
	edit := maps.Clone(own)
	edit["version"] = 2
	edit["photos"] = []map[string]any{{"id": published.Photos[0].ID}}
	codes := make(chan int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); codes <- rawRequest("PUT", private, edit, 1, true).Code }()
	}
	wg.Wait()
	close(codes)
	counts := map[int]int{}
	for code := range codes {
		counts[code]++
	}
	if counts[200] != 1 || counts[409] != 1 {
		t.Fatalf("edit conflict not protected: %v", counts)
	}
	// Reassignment removes access from the previous manager immediately.
	edit["version"] = 3
	edit["manager_id"] = accounts[2].ID
	request("PUT", adminPath, edit, 0, 200)
	request("GET", private, nil, 1, 404)
	request("GET", private, nil, 2, 200)
	delete(edit, "manager_id")
	edit["version"] = 4
	edit["published"] = false
	request("PUT", private, edit, 2, 200)
	request("GET", public, nil, -1, 404)
	request("GET", photoURL, nil, -1, 404)
	request("GET", fmt.Sprintf("%s/photos/%d", private, published.Photos[0].ID), nil, 1, 404)
	request("GET", fmt.Sprintf("%s/photos/%d", private, published.Photos[0].ID), nil, 2, 200)
	if total(request("GET", "/api/v1/me/favorites", nil, 3, 200)) != 0 {
		t.Fatal("hidden favorite leaked")
	}
	request("PUT", favoritePath, nil, 3, 404)
	edit["version"] = 5
	edit["manager_id"] = accounts[2].ID
	edit["status"] = "inactive"
	edit["published"] = false
	request("PUT", adminPath, edit, 0, 200)
	delete(edit, "manager_id")
	delete(edit, "status")
	edit["version"] = 6
	request("PUT", private, edit, 2, 403)
	request("DELETE", favoritePath, nil, 3, 200) // Hidden favorites can still be removed.
}
