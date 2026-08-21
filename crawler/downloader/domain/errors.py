"""Custom exception hierarchy for the downloader domain."""


class DownloaderError(Exception):
    """Base class for all downloader domain errors."""


class NotFoundError(DownloaderError):
    """Raised when an entity is not found. Non-retryable."""


class ValidationError(DownloaderError):
    """Raised when input fails domain validation. Non-retryable."""
