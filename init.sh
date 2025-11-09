#!/bin/bash

# Define o nome do binário/serviço
SERVICE_BIN="./tenant"
RUN_DIR="./run"

# --- 1. Garantir Permissão de Execução (chmod 777) ---
echo "=> Definindo permissão de execução total (777) para $SERVICE_BIN..."
chmod 777 "$SERVICE_BIN"
if [ $? -ne 0 ]; then
    echo "ERRO: Falha ao aplicar chmod em $SERVICE_BIN. O arquivo existe?"
    exit 1
fi
echo "=> Permissão concedida."

# --- 2. Parar o Serviço ---
echo "=> Tentando parar o serviço: $SERVICE_BIN --stop"
"$SERVICE_BIN" --stop
STOP_STATUS=$?

if [ $STOP_STATUS -eq 0 ]; then
    echo "=> Serviço parado com sucesso."
elif [ $STOP_STATUS -eq 127 ]; then
    echo "AVISO: Comando --stop não encontrado ou erro de permissão. Prosseguindo."
else
    # Assumimos que a falha ao parar é aceitável, pois pode não estar rodando
    echo "AVISO: O comando de parada retornou código $STOP_STATUS. Pode não estar rodando."
fi

sleep 2 # Dá um tempo para o sistema garantir a parada

# --- 3. Limpeza Agressiva do PID/RUN Dir ---
# Remove o diretório 'run' e todos os seus conteúdos, forçando a inicialização limpa.
if [ -d "$RUN_DIR" ]; then
    echo "=> LIMPANDO O CAOS: Removendo o diretório PID/run ($RUN_DIR) para garantir inicialização limpa..."
    rm -rf "$RUN_DIR"
    if [ $? -ne 0 ]; then
        echo "ERRO: Falha ao remover $RUN_DIR. Verifique permissões."
        exit 1
    fi
fi

# --- 4. Iniciar o Serviço ---
echo "=> Iniciando o serviço: $SERVICE_BIN --start"
"$SERVICE_BIN" --start
START_STATUS=$?

if [ $START_STATUS -ne 0 ]; then
    echo "ERRO FATAL: Falha ao iniciar o serviço. Código de saída: $START_STATUS"
    exit 1
else
    echo "SUCESSO: $SERVICE_BIN iniciado e rodando."
fi

# Fim do deploy
echo "--- DEPLOY CONCLUÍDO ---"