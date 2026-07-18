import sys

def contar_linhas_codigo(codigo_fonte: str) -> int:
    contador = 0
    # Divide o texto em uma lista de linhas
    linhas = codigo_fonte.split('\n')

    for linha in linhas:
        # Remove espaços nas extremidades para uma verificação precisa
        linha_limpa = linha.strip()

        # Aplica os filtros solicitados
        if not linha_limpa:
            continue  # Ignora linhas vazias (ou apenas com espaços)
            
        if linha_limpa.startswith(';'):
            continue  # Ignora linhas que começam com ';' (comentários)
            
        if linha_limpa.endswith(':'):
            continue  # Ignora linhas que terminam com ':' (labels)

        # Se a linha passou por todos os filtros, ela é contabilizada
        contador += 1

    return contador

def contar_linhas_arquivo(caminho_arquivo: str) -> int:
    with open(caminho_arquivo, 'r', encoding='utf-8') as arquivo:
        codigo_fonte = arquivo.read()
        return contar_linhas_codigo(codigo_fonte)

if __name__ == "__main__":
    # Verifica se o caminho do arquivo foi passado como argumento
    if len(sys.argv) < 2:
        print("Uso: python script.py <caminho_do_arquivo>")
        sys.exit(1)

    caminho_do_arquivo = sys.argv[1]
    
    try:
        total_linhas = contar_linhas_arquivo(caminho_do_arquivo)
        print(f"Total de linhas de código processadas: {total_linhas}")
    except FileNotFoundError:
        print(f"Erro: O arquivo '{caminho_do_arquivo}' não foi encontrado.")