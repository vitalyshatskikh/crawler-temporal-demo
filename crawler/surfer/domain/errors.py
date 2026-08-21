"""Custom exception hierarchy for the surfer domain."""


class SurferError(Exception):
    """Base class for all surfer domain errors."""


class NotFoundError(SurferError):
    """Raised when an entity is not found. Non-retryable."""


class ValidationError(SurferError):
    """Raised when input fails domain validation. Non-retryable."""
