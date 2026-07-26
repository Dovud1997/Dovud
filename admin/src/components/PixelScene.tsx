import { useEffect, useRef } from "react";
import type { Agent } from "../api";

type Props = {
  agents: Agent[];
};

type Zone = { id: string; label: string; x: number; y: number; w: number; h: number; color: string };

const ZONES: Zone[] = [
  { id: "telegram", label: "TELEGRAM", x: 40, y: 70, w: 240, h: 220, color: "#3d8bfd" },
  { id: "instagram", label: "INSTAGRAM", x: 320, y: 70, w: 240, h: 220, color: "#e1306c" },
  { id: "youtube", label: "YOUTUBE", x: 600, y: 70, w: 240, h: 220, color: "#ff4d4d" },
];

const STATUS_COLOR: Record<string, string> = {
  online: "#5ddea8",
  busy: "#f0c14a",
  error: "#ff6b6b",
  offline: "#8a93a5",
  draft: "#6b7280",
  connecting: "#7dd3fc",
};

export function PixelScene({ agents }: Props) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const agentsRef = useRef(agents);
  const frameRef = useRef(0);

  useEffect(() => {
    agentsRef.current = agents;
  }, [agents]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    let raf = 0;
    const draw = () => {
      frameRef.current += 1;
      const t = frameRef.current;
      ctx.imageSmoothingEnabled = false;

      // Atmospheric ground
      const grad = ctx.createLinearGradient(0, 0, 0, canvas.height);
      grad.addColorStop(0, "#1a2332");
      grad.addColorStop(0.45, "#243447");
      grad.addColorStop(1, "#14202c");
      ctx.fillStyle = grad;
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      // Pixel grid haze
      ctx.fillStyle = "rgba(255,255,255,0.03)";
      for (let y = 0; y < canvas.height; y += 8) {
        for (let x = 0; x < canvas.width; x += 8) {
          if ((x + y + t) % 24 === 0) ctx.fillRect(x, y, 2, 2);
        }
      }

      // Zones
      for (const zone of ZONES) {
        ctx.fillStyle = "rgba(0,0,0,0.28)";
        ctx.fillRect(zone.x + 4, zone.y + 4, zone.w, zone.h);
        ctx.fillStyle = `${zone.color}22`;
        ctx.fillRect(zone.x, zone.y, zone.w, zone.h);
        ctx.strokeStyle = zone.color;
        ctx.lineWidth = 3;
        ctx.strokeRect(zone.x, zone.y, zone.w, zone.h);

        // Zone sign
        ctx.fillStyle = zone.color;
        ctx.fillRect(zone.x + 12, zone.y - 18, zone.label.length * 9 + 16, 18);
        ctx.fillStyle = "#0b1220";
        ctx.font = "bold 12px 'Press Start 2P', monospace";
        ctx.fillText(zone.label, zone.x + 20, zone.y - 4);

        // Floor tiles
        ctx.fillStyle = "rgba(255,255,255,0.05)";
        for (let fx = zone.x + 8; fx < zone.x + zone.w - 8; fx += 16) {
          for (let fy = zone.y + zone.h - 40; fy < zone.y + zone.h - 8; fy += 16) {
            ctx.fillRect(fx, fy, 12, 12);
          }
        }
      }

      // Agents as pixel characters
      for (const agent of agentsRef.current) {
        const zone = ZONES.find((z) => z.id === agent.zone) || ZONES[0];
        const bob = agent.status === "busy" ? Math.sin(t / 6) * 3 : Math.sin(t / 18) * 1.5;
        const walk = agent.status === "busy" ? Math.sin(t / 5) * 6 : 0;
        const x = Math.min(Math.max(agent.pos_x, zone.x + 24), zone.x + zone.w - 40) + walk;
        const y = Math.min(Math.max(agent.pos_y, zone.y + 40), zone.y + zone.h - 50) + bob;
        drawCharacter(ctx, x, y, agent, t);
      }

      raf = requestAnimationFrame(draw);
    };

    raf = requestAnimationFrame(draw);
    return () => cancelAnimationFrame(raf);
  }, []);

  return (
    <canvas
      ref={canvasRef}
      width={880}
      height={360}
      className="pixel-canvas"
      aria-label="Pixel agent scene"
    />
  );
}

function drawCharacter(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  agent: Agent,
  t: number,
) {
  const color = STATUS_COLOR[agent.status] || "#aaa";
  const skin = "#f0d2b0";
  const shirt = agent.platform === "telegram" ? "#3d8bfd" : agent.platform === "instagram" ? "#e1306c" : "#ff4d4d";

  // shadow
  ctx.fillStyle = "rgba(0,0,0,0.35)";
  ctx.fillRect(x + 4, y + 34, 20, 6);

  // legs
  ctx.fillStyle = "#2a3344";
  const legSwing = agent.status === "busy" ? Math.sin(t / 4) * 3 : 0;
  ctx.fillRect(x + 6, y + 26, 5, 10 + legSwing);
  ctx.fillRect(x + 15, y + 26, 5, 10 - legSwing);

  // body
  ctx.fillStyle = shirt;
  ctx.fillRect(x + 4, y + 12, 20, 16);

  // head
  ctx.fillStyle = skin;
  ctx.fillRect(x + 6, y, 16, 14);
  ctx.fillStyle = "#1a1a1a";
  ctx.fillRect(x + 9, y + 5, 3, 3);
  ctx.fillRect(x + 16, y + 5, 3, 3);

  // status halo
  ctx.fillStyle = color;
  ctx.fillRect(x + 8, y - 8, 12, 4);
  if (agent.status === "error" && Math.floor(t / 10) % 2 === 0) {
    ctx.fillStyle = "#ff6b6b";
    ctx.fillRect(x + 22, y - 16, 8, 8);
  }
  if (agent.status === "busy") {
    ctx.fillStyle = "#f0c14a";
    ctx.fillRect(x + 26, y + 8, 6, 6);
  }

  // nameplate
  ctx.fillStyle = "rgba(10,16,28,0.75)";
  ctx.fillRect(x - 10, y + 40, 48, 14);
  ctx.fillStyle = "#e8eef8";
  ctx.font = "10px 'IBM Plex Mono', monospace";
  const label = agent.name.slice(0, 8);
  ctx.fillText(label, x - 6, y + 50);
}
