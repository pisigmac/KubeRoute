"""Kuberoute SDK Exceptions."""

from typing import Optional, Any


class KubeRouteError(Exception):
    """Base exception for all KubeRoute SDK errors."""

    def __init__(self, message: str, status_code: Optional[int] = None, details: Optional[Any] = None) -> None:
        super().__init__(message)
        self.message = message
        self.status_code = status_code
        self.details = details

    def __str__(self) -> str:
        if self.status_code:
            return f"[{self.status_code}] {self.message}"
        return self.message


class AuthenticationError(KubeRouteError):
    """Raised when authentication fails (401 Unauthorized or 403 Forbidden)."""
    pass


class NotFoundError(KubeRouteError):
    """Raised when a requested resource is not found (404)."""
    pass


class ValidationError(KubeRouteError):
    """Raised when request payload or parameters fail validation (400 Bad Request)."""
    pass


class ServerError(KubeRouteError):
    """Raised when the Kuberoute server returns an internal server error (500+)."""
    pass
