import datetime as dt
import logging
import os
import typing as tp

import pydantic
import pydantic_settings
from temporalio import common

from shared.py.db import settings as db_settings


class Config(pydantic_settings.BaseSettings):
    model_config = pydantic_settings.SettingsConfigDict(
        env_file=("conf/.env", os.getenv("ENV_FILE", ".env")),
        env_nested_delimiter="__",
        env_file_encoding="utf-8",
        enable_decoding=False,
        nested_model_default_partial_update=True,
        populate_by_name=True,
        extra="ignore",
    )

    app_name: str = pydantic.Field("my-app")
    app_version: str = pydantic.Field("0.1.0")
    app_environment: str = pydantic.Field("development")

    logging_level: int = pydantic.Field(logging.INFO)
    logging_add_caller: bool = pydantic.Field(False)

    temporal_host: str = pydantic.Field("localhost:7233")
    temporal_namespace: str = pydantic.Field("crawler")

    postgres: db_settings.PGConfig = db_settings.PGConfig()


class RetryConfig(pydantic.BaseModel):
    max_attempts: int = pydantic.Field(3, ge=0)
    init_interval: dt.timedelta = pydantic.Field(dt.timedelta(seconds=1), ge=dt.timedelta(0))
    max_interval: dt.timedelta = pydantic.Field(dt.timedelta(seconds=30), ge=dt.timedelta(0))
    backoff_coefficient: float = pydantic.Field(2.0, ge=0.0)

    def to_retry_policy(
        self,
        *,
        non_retryable_error_types: tp.Sequence[str] = (),
    ) -> common.RetryPolicy:
        return common.RetryPolicy(
            maximum_attempts=self.max_attempts,
            initial_interval=self.init_interval,
            maximum_interval=self.max_interval,
            backoff_coefficient=self.backoff_coefficient,
            non_retryable_error_types=non_retryable_error_types,
        )
