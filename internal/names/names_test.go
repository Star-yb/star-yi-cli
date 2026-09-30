package names

import "testing"

func TestCleanAppsDir(t *testing.T) {
	ok := map[string]string{
		"apps":          "apps",
		" apps/ ":       "apps",
		`modules\biz`:   "modules/biz",
		"./apps":        "apps",
		"biz/./modules": "biz/modules",
	}
	for in, want := range ok {
		got, err := CleanAppsDir(in)
		if err != nil || got != want {
			t.Errorf("CleanAppsDir(%q) = %q, %v；期望 %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", ".", "..", "../x", "/abs", `C:\x`, "yi-admin", "yi-common/x", "a b", ".git"} {
		if _, err := CleanAppsDir(in); err == nil {
			t.Errorf("CleanAppsDir(%q) 应该报错", in)
		}
	}
}

func TestModuleAndPackage(t *testing.T) {
	for _, id := range []string{"order", "order-service", "a1"} {
		if err := ValidateModuleID(id); err != nil {
			t.Errorf("%s: %v", id, err)
		}
	}
	for _, id := range []string{"", "Order", "order-", "a--b", "1a", "yi-admin", "a/b", "a.b"} {
		if ValidateModuleID(id) == nil {
			t.Errorf("%q 应该不合法", id)
		}
	}
	if DefaultPackage("order-service") != "orderservice" || DefaultPackage("object") != "objectapp" {
		t.Error("DefaultPackage")
	}
	if ClassName("order-service") != "OrderService" {
		t.Error("ClassName")
	}
	if ValidatePackage("order.core") != nil || ValidatePackage("class") == nil || ValidatePackage("Order") == nil {
		t.Error("ValidatePackage")
	}
}
