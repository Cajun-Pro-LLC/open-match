package config

// Matchmaker Uses a JSON/YAML configuration to configure your matchmaker, so you don't have to.
// Modeled after the Edgegap No-code Managed Matchmaker
type Matchmaker struct {
	// Configuration for the authentication method of your matchmaker
	Auth *MatchmakerAuth `json:"auth"`
	// A List of profiles for your matchmaker. (must contain at least 1 element)
	Profiles []*MatchmakerProfile `json:"profiles"`
}

// MatchmakerAuth Configuration for the authentication method of your matchmaker
type MatchmakerAuth struct {
	// Type of authentication used by the Front End (possible values are -> NoAuth|Token)
	Type string `json:"type"`
	// Configuration of authentication's type. (optional)
	Configuration *MatchmakerAuthConfiguration `json:"configuration"`
}

// MatchmakerAuthConfiguration Configuration of authentication's type. (optional)
type MatchmakerAuthConfiguration struct {
	// Token used to authenticate the request. The token will be fetched from the headers at key authorization.
	//
	// Note: This is only used in Token authentication type
	Token string `json:"token"`
	// This is not used for the moment (optional)
	Key string `json:"key"`
}

// MatchmakerProfile A profile for your matchmaker
type MatchmakerProfile struct {
	// A unique ID to identify the profile.
	//
	// Constraint: globally unique
	Id string `json:"profile_id"`
	// The displayable name of the profile
	Name string `json:"name"`
	// The arbitrium application to deploy by this profile
	App string `json:"app"`
	// The version of your application that will be deployed
	Version string `json:"version"`
	// The port name you assigned to the port in your application version, which is used by the player to connect
	GamePort string `json:"game_port"`
	// The delay in seconds before your matchmaker start
	//
	// Recommended value range: 2-10
	DelayToStart int32 `json:"delay_to_start"`
	// The time in seconds between each iteration of matchmaking.
	//
	// Recommended value range: 5-10 - Any values under 5 will be ignored
	Refresh int32 `json:"refresh"`
	// The number of player in a match
	MatchPlayerCount int32 `json:"match_player_count"`
	// Whether the values are injected into the deployment - optional
	InjectEnv bool `json:"inject_env"`
	// A list of selectors. They represent a choice between discreet values.
	// Only players with the same choices in all selectors will be matched together.
	Selectors []*MatchmakerSelector `json:"selectors"`
	// A list of range values. They represent a choice between 2 numerical values.
	// Only players within the same range will be matched together. (Only support 1 filter at the moment)
	Filters []*MatchmakerFilter `json:"filters"`
}

// MatchmakerSelector represents a choice between discreet values.
// Only players with the same choices in all selectors will be matched together.
type MatchmakerSelector struct {
	// A unique ID to identify the selector.
	//
	// Constraint: profile unique
	Key string `json:"key"`
	// The displayable name of the selector
	Name string `json:"name"`
	// The default value for this selector - optional
	Default string `json:"default"`
	// Whether the client is required to send data for this selector.
	// If the selector is required and the player doesn't send data, it will deny the player's request.
	Required bool `json:"required"`
	// Whether the values are injected into the deployment - optional
	InjectEnv bool `json:"inject_env"`
	// A list of possible values for this selector. Must contain at least 1 element.
	Items []string `json:"items"`
}

// MatchmakerFilter represents range values.
// Only players within the same range will be matched together.
type MatchmakerFilter struct {
	// A unique ID to identify the filter
	//
	// Constraint: profile unique
	Key string `json:"key"`
	// The displayable name of the filter
	Name string `json:"name"`
	// Whether the client is required to send data for this filter.
	// If the filter is required and the player doesn't send data, it will deny the player's request.
	Required bool `json:"required"`
	// Maximum value of the range - inclusive
	Maximum float64 `json:"maximum"`
	// Minimum value of the range - inclusive
	Minimum float64 `json:"minimum"`
	// Configuration of the range
	Difference *MatchmakerFilterDifference `json:"difference"`
}

// MatchmakerFilterDifference Configuration of the MatchmakerFilter range
type MatchmakerFilterDifference struct {
	// Negative value of the range. If the player value is 200 and the negative value is 40, he
	// can be matched with players that have a value down to 160 inclusive.
	Negative float64 `json:"negative"`
	// Positive value of the range. If the player value is 200 and the positive value is 40, he
	// can be matched with players that have a value up to 240 inclusive.
	Positive float64 `json:"positive"`
}
