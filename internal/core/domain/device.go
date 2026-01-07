package domain

// DeviceType 定义设备类型
type DeviceType string

const (
	DeviceTypeWater DeviceType = "WATER" // 水表
	DeviceTypeElec  DeviceType = "ELEC"  // 电表
	DeviceTypeGas   DeviceType = "GAS"   // 燃气表
	DeviceTypeHeat  DeviceType = "HEAT"  // 热量表
)

// DeviceInfo 包含设备的静态属性
// 对应需求 3.2: 设备信息（型号、类型）
type DeviceInfo struct {
	ID    string     `json:"device_id"`
	Model string     `json:"model"`
	Type  DeviceType `json:"type"`
}
