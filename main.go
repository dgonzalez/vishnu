package main

func main() {
	vishu := New(func(stats map[string]interface{}) int {
		if stats["timeout"] != nil {
			return stats["timeout"].(int)
		}
		return 0
	})
	vishu.Add("test")
	vishu.Add("test2")
	for i := 0; i < 10; i++ {
		vishu.With(func(endpoint interface{}) (map[string]interface{}, error) {
			stats := make(map[string]interface{})
			endpointStr := endpoint.(string)
			if endpointStr == "test2" {
				stats["timeout"] = 1000
			}
			return stats, nil
		})
	}
}
