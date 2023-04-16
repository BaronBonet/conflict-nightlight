from app.core import domain
from generated.conflict_nightlight.v1 import Map, Date, MapProvider, MapType, Bounds


def transform_map_domain_to_proto(m: domain.Map) -> Map:
    return Map(
        date=Date(day=m.date.day, month=m.date.month, year=m.date.year),
        map_type=m.map_type,
        map_source=m.map_source,
        bounds=m.bounds,
    )


def map_provider_to_string(p: MapProvider) -> str | None:
    match p:
        case MapProvider.MAP_PROVIDER_EOGDATA:
            return "Eogdata"
    return None


def map_type_to_string(t: MapType) -> str | None:
    match t:
        case MapType.MAP_TYPE_MONTHLY:
            return "Monthly"
        case MapType.MAP_TYPE_DAILY:
            return "Daily"
    return None


def bounds_to_string(b: Bounds) -> str | None:
    match b:
        case Bounds.BOUNDS_UKRAINE_AND_AROUND:
            return "UkraineAndAround"
    return None
