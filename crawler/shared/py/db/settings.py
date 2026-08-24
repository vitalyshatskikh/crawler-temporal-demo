import urllib.parse

import pydantic


class PGConfig(pydantic.BaseModel):
    hosts: str = pydantic.Field("localhost:5432")
    user: str = pydantic.Field("postgres")
    password: str = pydantic.Field("postgres")
    database: str = pydantic.Field("crawler")

    def to_dsn(self) -> str:
        return (
            f"postgresql+asyncpg://{self.user}:{urllib.parse.quote(self.password)}"
            f"@{self.hosts}/{self.database}"
        )
