import pathlib
from abc import ABC, abstractmethod

from app.core import domain


class ExternalMapRepository(ABC):
    """Driven port defining the interface for external maps"""

    @abstractmethod
    def download(self, m: domain.Map) -> domain.LocalMap:
        pass


class InternalMapRepository(ABC):
    """Driven port defining the interface for internal maps"""

    @abstractmethod
    def save(self, m: domain.LocalMap):
        pass

    @abstractmethod
    def download(self, m: domain.Map) -> domain.LocalMap:
        pass


class BoundsRepository(ABC):
    """Driven port defining the interface for our boundary files"""

    @abstractmethod
    def download(self, key: str) -> pathlib.PosixPath | None:
        """
        Downloads all shapefiles required

        Returns:
            The location on the local file system of the boundary file
        """


class Logger(ABC):
    """Driven port defining the interface for our logger"""

    def debug(self, message: str, **kwargs):
        pass

    def info(self, message: str, **kwargs):
        pass

    def warning(self, message: str, **kwargs):
        pass

    def error(self, message: str, **kwargs):
        pass

    def fatal(self, message: str, **kwargs):
        pass
