package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/slaghuis/YC-Image/pkg/config"
)

// UserRole type to represent different roles
type UserRole string

const (
	RoleVisitor    UserRole = "visitor"
	RoleCustomer   UserRole = "customer"
	RoleAdmin      UserRole = "admin"
  RoleSysAdmin   UserRole = "sysadmin"
)

// RBACMiddleware is a middleware to enforce role-based access control
func RBACMiddleware(requiredRoles ...UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the user role from the context (set by a previous authentication middleware)
		//    In a real app, this would be retrieved from a validated JWT or session
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort() // Abort prevents pending handlers from being called
			return
		}

    currentUserRole := UserRole(userRole.(string))

		// 2. Check if the user's role is in the list of required roles
		hasPermission := false
		for _, role := range requiredRoles {
			if currentUserRole == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// 3. If permitted, continue to the next handler
		c.Next()
	}
}


func RequireAccessToken(cfg config.JwtConfigurations) gin.HandlerFunc {
    return func(c *gin.Context) {
        auth := c.GetHeader("Authorization")
        if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
            return
        }
        tok := strings.TrimSpace(auth[len("Bearer "):])
        claims, err := parseAccessToken(cfg, tok)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        c.Set("user_id", claims.Sub)
        c.Set("user_role", claims.Role)
        c.Set("user_name", claims.Name)
        c.Set("user.email", claims.Email)

        c.Next()
    }
}
