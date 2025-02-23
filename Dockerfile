FROM golang:1.24.0-alpine

# 기본 설정
WORKDIR /app

# 필요한 패키지 설치 (Air 포함)
RUN apk add --no-cache git \
    && go install github.com/air-verse/air@latest

# 의존성 설치
COPY go.mod go.sum ./
RUN go mod download

# 코드 복사
COPY . .

# entrypoint.sh 복사
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# 컨테이너 실행 시 entrypoint.sh 실행
ENTRYPOINT ["/entrypoint.sh"]

# 기본 실행 명령 (air 사용)
CMD ["air", "-c", ".air.toml"]
