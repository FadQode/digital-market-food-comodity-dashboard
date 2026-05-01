import os
from dotenv import load_dotenv
import psycopg2

load_dotenv()

def run_sql_file(path):
    conn = psycopg2.connect(
        host=os.getenv("DB_HOST"),
        port=os.getenv("DB_PORT"),
        dbname=os.getenv("DB_NAME"),
        user=os.getenv("DB_USER"),
        password=os.getenv("DB_PASSWORD"),
        sslmode=os.getenv("DB_SSLMODE")
    )
    
    cur = conn.cursor()
    
    with open(path, 'r') as f:
        sql = f.read()
        cur.execute(sql)
    
    conn.commit()
    cur.close()
    conn.close()
    
    print(f"✅ Executed: {path}")