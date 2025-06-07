echo "[CLEAN] Stopping containers..."
docker compose down --volumes --remove-orphans

echo "[CLEAN] Removing all images..."
docker rmi -f $(docker images -q)

echo "[CLEAN] Pruning builder cache and system..."
docker system prune -a -f --volumes
docker builder prune -a -f

echo "[CLEAN] Done. Disk usage now:"
df -h /

