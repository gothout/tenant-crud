CREATE TABLE tenant (
    -- Chave Primária e Identificador Único Universal (UUID)
                        uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Nome do Tenant (não pode ser nulo)
                        name VARCHAR(255) NOT NULL,

    -- Documento (CNPJ/CPF, etc.). Deve ser único na tabela e não pode ser nulo.
                        document VARCHAR(100) NOT NULL UNIQUE,

    -- Status de Atividade (booleano, padrão é 'true')
                        live BOOLEAN NOT NULL DEFAULT TRUE,

    -- Data e Hora da Criação - PASSADA PELO BACKEND
                        create_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,

    -- Data e Hora da Última Atualização - PASSADA PELO BACKEND
                        update_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

-- Opcional: Criação de um índice para o campo 'document' para otimizar pesquisas
CREATE INDEX idx_tenant_document ON tenant (document);