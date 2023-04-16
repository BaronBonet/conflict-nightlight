import tarfile
import pathlib
import uuid

import requests

from app.core import ports, domain
from app.core.ports import ExternalMapRepository
from app.infrastructure.get_or_raise import get_or_raise

TOKEN_URL = "https://eogauth.mines.edu/auth/realms/master/protocol/openid-connect/token"


class EogdataMapRepository(ExternalMapRepository):
    def __init__(self, local_write_directory: pathlib.Path, logger: ports.Logger, secrets: dict[str, str]):
        logger.debug("Initiating EogdataMapRepository")
        self.local_write_directory = local_write_directory
        self.logger = logger

        username = get_or_raise(secrets, "eogdataUsername")
        password = get_or_raise(secrets, "eogdataPassword")

        self.token = self._get_eog_auth_token(TOKEN_URL, username, password)

    def download(self, m: domain.Map) -> domain.LocalMap:
        self.logger.info("Starting download", download_url=m)
        response = requests.get(m.map_source.url, headers={"Authorization": f"Bearer {self.token}"}, stream=True)

        output_file_name = pathlib.Path(self.local_write_directory) / f"{uuid.uuid4()}.tgz"

        with open(output_file_name, "wb") as f:
            for chunk in response.iter_content(1024):
                if chunk:
                    f.write(chunk)

        return domain.LocalMap(map=m, file_path=self._extract_tif(output_file_name))

    def _extract_tif(self, file_to_unzip: pathlib.Path) -> pathlib.Path:
        tar = tarfile.open(str(file_to_unzip))
        extracted_file = ""
        for item in tar:
            if {*pathlib.Path(item.name).suffixes} == {".avg_rade9h", ".tif"}:
                tar.extract(item, path=self.local_write_directory)
                extracted_file = item.name

        if not extracted_file:
            raise Exception(
                "There was no file that contained the expected suffixes",
            )
        return file_to_unzip.parent / extracted_file

    def _get_eog_auth_token(self, token_url: str, username: str, password: str) -> str:
        """
        Get an EOG auth token for the API requests

        Instructions from: https://eogdata.mines.edu/products/register/
        """
        params = {
            "client_id": "eogdata_oidc",
            "client_secret": "2677ad81-521b-4869-8480-6d05b9e57d48",
            "username": username,
            "password": password,
            "grant_type": "password",
        }
        try:
            token_resp = requests.post(token_url, data=params)
        except requests.exceptions.ConnectionError as e:
            self.logger.fatal(
                "There was a connection error while connecting to eogdata when attempting to retrieve the token",
                token_url=token_url,
                error=e,
            )

        if e := token_resp.json().get("error"):
            self.logger.fatal(
                "There was an error returned from the Eogdata token server",
                token_resp=token_resp,
                error=e,
            )

        token = token_resp.json().get("access_token")
        if not token:
            self.logger.fatal("The access token does not exist in the token object")
        return token
