import math
import pathlib

import numpy as np
import rasterio

from app.core import domain
from app.core.ports import InternalMapRepository


class MapProductService:
    def __init__(self, raw_map_repository: InternalMapRepository, processed_map_repository: InternalMapRepository):
        self.raw_map_repository = raw_map_repository
        self.processed_map_repository = processed_map_repository

    def process_save(self, m: domain.Map):
        local_map = self.raw_map_repository.download(m)
        local_processed_file = _process_tif(local_map.file_path, lower_clip=0, upper_clip=2000)
        self.processed_map_repository.save(domain.LocalMap(map=m, file_path=local_processed_file))


def _get_raster_data(raster_file: str):
    r = rasterio.open(raster_file)
    return r.read()[0]


def _create_out_file_name(input_file: pathlib.PurePath) -> pathlib.PurePath:
    file_name = input_file.name.split(".")[0]
    return input_file.parent / file_name


def _process_tif(raw_tif: pathlib.Path, lower_clip: int, upper_clip: int) -> pathlib.Path:
    out_file_name = _create_out_file_name(raw_tif)
    processed_image_name = pathlib.Path(f"{out_file_name}_processed.tif")

    with rasterio.open(raw_tif) as src:
        data = src.read()[0]

        data = np.clip(data, lower_clip, upper_clip)

        data = np.log10(data + 1)

        # Rescale the data to a range of 0 to 255, to ensure it is a 8bit int, that way it can be uploaded to mapbox.
        data = np.interp(data, (lower_clip, math.log10(upper_clip)), (0, 255))

        with rasterio.open(
            processed_image_name,
            "w",
            driver="GTiff",
            height=src.height,
            width=src.width,
            count=1,
            dtype="uint8",
            crs=src.crs,
            transform=src.transform,
        ) as dst:
            dst.write(data, 1)

    return processed_image_name
