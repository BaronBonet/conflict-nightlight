import datetime
import pathlib
from dataclasses import dataclass
from generated.conflict_nightlight import v1


@dataclass
class Map:
    date: datetime.date
    map_type: v1.MapType
    map_source: v1.MapSource
    bounds: v1.Bounds


@dataclass
class LocalMap:
    map: Map
    file_path: pathlib.Path
