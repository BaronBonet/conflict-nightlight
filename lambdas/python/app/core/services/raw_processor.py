import pathlib

import geopandas as gpd
import rasterio.mask
from shapely.geometry import mapping

from app.core import domain, ports
from app.core.ports import BoundsRepository, ExternalMapRepository, InternalMapRepository


class RawMapProcessorService:
    def __init__(
        self,
        external_map_repository: ExternalMapRepository,
        internal_map_repository: InternalMapRepository,
        bounds_repository: BoundsRepository,
        logger: ports.Logger,
    ):
        self.external_map_repository = external_map_repository
        self.internal_map_repository = internal_map_repository
        self.bounds_repository = bounds_repository
        self.logger = logger

    def download_crop_save(self, m: domain.Map):
        boundary_file = self.bounds_repository.download(m.bounds.name)
        raw_internal_map = self.external_map_repository.download(m=m)
        cropped_map_location = self._crop_raw_file(raw_internal_map.file_path, boundary_file)
        self.internal_map_repository.save(domain.LocalMap(file_path=cropped_map_location, map=m))

    def _crop_raw_file(self, raw_map_file: pathlib.Path, boundary_file: pathlib.Path) -> pathlib.Path:
        self.logger.debug("Starting to crop file", raw_map_file=raw_map_file, boundary_file=boundary_file)
        crop_boundaries = self._get_crop_boundaries(boundary_file)

        with rasterio.open(str(raw_map_file)) as raster_data:
            cropped_data, crop_data_afflin = rasterio.mask.mask(
                dataset=raster_data,
                shapes=[crop_boundaries],
                crop=True,
                filled=False,
            )
            meta_data = raster_data.meta

        meta_data.update(
            {
                "transform": crop_data_afflin,
                "height": cropped_data.shape[1],
                "width": cropped_data.shape[2],
                "nodata": 0,
            }
        )
        cropped_file_name = raw_map_file.parent / f"{raw_map_file.name}_cropped_{raw_map_file.suffix}"

        with rasterio.open(
            cropped_file_name,
            "w",
            **meta_data,
        ) as out_file:
            out_file.write(cropped_data)
            self.logger.debug("Sucessfully cropped file", out_file=out_file)
        return cropped_file_name

    def _get_crop_boundaries(self, boundary_shp_file: pathlib.Path) -> dict:
        """
        Extract the crop boundaries from a shape file.

        Args:
            boundary_shp_file (str): The path to the shape file containing the crop boundaries.

        Returns:
            dict: A dictionary of coordinates representing the crop boundaries.
        """
        self.logger.debug("Extracting crop boundaries", boundary_shp_file=boundary_shp_file)
        crop_extent = gpd.read_file(boundary_shp_file)
        return mapping(crop_extent["geometry"][0])
