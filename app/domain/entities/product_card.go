package entities

type ProductCard struct {
	ProductTitle         string
	ProductDescription   string
	ParentId             int32
	SubjectId            int32
	RootId               int32
	SubId                int32
	TypeId               int32
	GenerateContent      bool
	Ozon                 bool
	Wb                   bool
	Translate            bool
	VendorCode           string
	Dimensions           *WBDimensions
	Brand                string
	Sizes                []*WBSize
	WbApiKey             string
	WbMediaToUploadFiles []*WBClientMediaFile
	WbMediaToSaveLinks   []string
	OzonApiClientId      string
	OzonApiKey           string
}

func (pc *ProductCard) GetOzonApiClientId() string {
	return pc.OzonApiClientId
}

func (pc *ProductCard) GetOzonApiKey() string {
	return pc.OzonApiKey
}

func (pc *ProductCard) GetWbApiKey() string {
	return pc.WbApiKey
}

func (pc *ProductCard) GetVendorCode() string {
	return pc.VendorCode
}

func (pc *ProductCard) GetOzon() bool {
	return pc.Ozon
}

func (pc *ProductCard) GetWb() bool {
	return pc.Wb
}

func (pc *ProductCard) GetWbMediaToUploadFiles() []*WBClientMediaFile {
	return pc.WbMediaToUploadFiles
}

func (pc *ProductCard) GetWbMediaToSaveLinks() []string {
	return pc.WbMediaToSaveLinks
}
