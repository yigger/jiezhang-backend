package config

// Typed string aliases for dig dependency injection.
// Plain string params would collide in the container;
// named types let dig distinguish them.

type TokenSecret string
type PublicBaseURL string
type MiniProgramAppID string
type MiniProgramSecret string
type MySQLDSN string
