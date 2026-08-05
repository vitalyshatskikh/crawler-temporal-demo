import datetime as dt
import logging
import sys
import traceback
import typing as tp

from pythonjsonlogger.json import JsonFormatter

from shared.py import settings


class _JsonFormatter(JsonFormatter):
    def __init__(self, fmt: tp.Any, include_caller: tp.Any, **kwargs: tp.Any) -> None:
        super().__init__(fmt=fmt, **kwargs)
        self._include_caller = include_caller

    def add_fields(
        self,
        log_record: dict[str, tp.Any],
        record: logging.LogRecord,
        message_dict: dict[str, tp.Any],
    ) -> None:
        super().add_fields(log_record, record, message_dict)
        log_record["timestamp"] = dt.datetime.fromtimestamp(record.created).isoformat()  # noqa: DTZ006
        log_record["level"] = record.levelname.lower()
        log_record["logger"] = record.name
        if self._include_caller:
            log_record["caller"] = f"{record.filename}:{record.lineno}"
        if record.exc_info:
            log_record["stacktrace"] = "".join(
                traceback.format_exception(*record.exc_info)
            )
        log_record.pop("levelname", None)
        log_record.pop("name", None)
        log_record.pop("exc_info", None)


def setup_logger(conf: settings.Config) -> None:
    root_logger = logging.getLogger()
    root_logger.handlers.clear()

    fmt_fields = ["timestamp", "level", "logger", "message"]
    if conf.logging_add_caller:
        fmt_fields.append("caller")
    fmt = " ".join(f"%({f})s" for f in fmt_fields)

    formatter = _JsonFormatter(
        fmt=fmt,
        include_caller=conf.logging_add_caller,
        datefmt="%Y-%m-%dT%H:%M:%S.%f%z",
    )

    handler = logging.StreamHandler(stream=sys.stdout)
    handler.setFormatter(formatter)
    root_logger.addHandler(handler)
    root_logger.setLevel(conf.logging_level)
