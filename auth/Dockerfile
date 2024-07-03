FROM golang:1.23rc1-alpine

# Definir o diretório de trabalho dentro do container
WORKDIR /app

# Copiar o arquivo go.mod e go.sum para o container e baixar as dependências
COPY go.mod go.sum ./
RUN go mod tidy

# Copiar o código fonte para o container
COPY . .

# Instalar o Service Weaver
RUN go install github.com/ServiceWeaver/weaver/cmd/weaver@latest

RUN weaver generate .

# Compilar o código Go
RUN go build -o auth .

# Expõe a porta 8080, onde a aplicação estará escutando
EXPOSE 8080

ENV SERVICEWEAVER_CONFIG weaver.toml

# Comando para iniciar a aplicação
CMD ["go", "run", "."]
