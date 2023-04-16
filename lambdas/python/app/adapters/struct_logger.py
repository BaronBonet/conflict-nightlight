import logging.config
import logging
import sys

import structlog
from structlog import get_logger
from inspect import currentframe, getframeinfo

from app.core.ports import Logger


def add_caller_logger_name(logger, method_name, event_dict):
    frame = currentframe()
    # Set the number of stack frames to skip since this is now an implementation of a port.
    caller_frame = frame.f_back.f_back.f_back.f_back.f_back.f_back
    caller_info = getframeinfo(caller_frame)
    event_dict["logger"] = f"{caller_info.filename}:{caller_info.lineno}"
    return event_dict


PROCESSORS = [
    structlog.contextvars.merge_contextvars,
    add_caller_logger_name,
    structlog.stdlib.add_log_level,
    structlog.stdlib.PositionalArgumentsFormatter(),
    structlog.processors.TimeStamper(fmt="iso"),
    structlog.processors.format_exc_info,
    structlog.processors.UnicodeDecoder(),
]


class StructLogger(Logger):
    def __init__(self, version: str, use_debug: bool, noisy_logs: list[str], correlation_id: str):
        log_level = logging.INFO
        if use_debug:
            log_level = logging.DEBUG

        structlog.contextvars.bind_contextvars(version=version, correlation_id=correlation_id)
        structlog.configure(
            context_class=dict,
            logger_factory=structlog.stdlib.LoggerFactory(),
            cache_logger_on_first_use=True,
            processors=PROCESSORS + [structlog.stdlib.ProcessorFormatter.wrap_for_formatter],
        )
        formatter = structlog.stdlib.ProcessorFormatter(
            foreign_pre_chain=PROCESSORS,
            processors=[
                structlog.stdlib.ProcessorFormatter.remove_processors_meta,
                structlog.processors.JSONRenderer(),
            ],
        )
        handler = logging.StreamHandler()
        handler.setFormatter(formatter)

        logger = logging.getLogger()
        logger.addHandler(handler)
        logger.setLevel(log_level)
        for source in noisy_logs:
            logging.getLogger(source).setLevel(logging.WARNING)

        self.logger = get_logger()

    def debug(self, msg: str, **kwargs):
        self.logger.debug(msg, **kwargs)

    def info(self, msg: str, **kwargs):
        self.logger.info(msg, **kwargs)

    def warning(self, msg: str, **kwargs):
        self.logger.warning(msg, **kwargs)

    def error(self, msg: str, **kwargs):
        self.logger.error(msg, **kwargs)

    def fatal(self, msg: str, **kwargs):
        self.logger.error(msg, **kwargs)
        sys.exit(msg)
