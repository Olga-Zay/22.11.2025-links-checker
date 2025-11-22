package dto

type CheckLinksRequest struct {
	Links []string `json:"links"`
}

type CheckLinksResponse struct {
	Links    map[string]string `json:"links"`
	LinksNum int64             `json:"links_num"`
}

type GenerateReportRequest struct {
	LinksList []int64 `json:"links_list"`
}
