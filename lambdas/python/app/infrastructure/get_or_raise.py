from typing import Any


def get_or_raise(dictionary: dict[str, Any], key: str):
    value = dictionary.get(key)
    if not value:
        raise Exception(f"No value was found when trying to download {key}")
    return value
