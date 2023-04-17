import datetime

from app.core import domain
from generated.conflict_nightlight.v1 import Map, Date, MapProvider, MapType, Bounds, MapSource


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
            return "MapProviderEogdata"
    return None


def map_type_to_string(t: MapType) -> str | None:
    match t:
        case MapType.MAP_TYPE_MONTHLY:
            return "MapTypeMonthly"
        case MapType.MAP_TYPE_DAILY:
            return "MapTypeDaily"
    return None


def bounds_to_string(b: Bounds) -> str | None:
    match b:
        case Bounds.BOUNDS_UKRAINE_AND_AROUND:
            return "BoundsUkraineAndAround"
    return None


def proto_map_to_domain(m: Map) -> domain.Map:
    return domain.Map(
        date=datetime.date(m.date.year, m.date.month, m.date.day),
        map_type=MapType(m.map_type),
        map_source=MapSource(map_provider=MapProvider(m.map_source.map_provider), url=m.map_source.url),
        bounds=Bounds(m.bounds),
    )
