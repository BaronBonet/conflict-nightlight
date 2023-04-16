ARG FUNCTION_DIR="/home/app"

FROM python:3.10.6-bullseye AS base

# lambda dependencies
RUN apt-get update && \
  apt-get install -y \
  g++ \
  make \
  cmake \
  unzip \
  libcurl4-openssl-dev

# Gdal dependencies
ARG CPLUS_INCLUDE_PATH=/usr/include/gdal
ARG C_INCLUDE_PATH=/usr/include/gdal

RUN apt-get -y update
RUN apt-get install -y software-properties-common && apt-get update
RUN apt-get install -y gdal-bin libgdal-dev

# Clean up build
RUN apt-get -y auto-remove \
    && apt-get -y clean \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /usr/share/doc/ \
    && rm -rf /usr/share/man/ \
    && rm -rf /usr/share/locale/

FROM base AS python-dependencies

ARG FUNCTION_DIR

ENV \
    POETRY_NO_INTERACTION=1 \
    PATH="/root/.local/bin:${PATH}" \
    POETRY_VERSION=1.3.1

WORKDIR ${FUNCTION_DIR}/

COPY poetry.lock pyproject.toml ${FUNCTION_DIR}/

RUN curl -sSL https://install.python-poetry.org | python3 -

RUN poetry export --without-hashes --format=requirements.txt > requirements.txt

RUN python3.10 -m pip install -r requirements.txt --target ${FUNCTION_DIR}/

FROM base

ARG VERSION="unknown"
ARG APP_DIRECTORY_NAME
ARG FUNCTION_DIR

ENV VERSION=$VERSION

WORKDIR ${FUNCTION_DIR}

COPY --from=python-dependencies ${FUNCTION_DIR}/ ${FUNCTION_DIR}/
COPY generated ${FUNCTION_DIR}/generated
COPY app ${FUNCTION_DIR}/app
COPY lambda_handler.py ${FUNCTION_DIR}/

# So we can emulate the lambda runtime on our local machine
ADD https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie /usr/bin/aws-lambda-rie
COPY scripts/entry.sh /
RUN chmod 755 /usr/bin/aws-lambda-rie /entry.sh

ENTRYPOINT [ "/entry.sh" ]
CMD [ "lambda_handler.handler" ]
