<script>
  import { onMount, onDestroy } from 'svelte';

  let canvas;
  let ctx;
  let raf = 0;
  let particles = [];
  let width = 0;
  let height = 0;

  const COUNT = 70;
  const palette = ['#4f8cff', '#7aa7ff', '#3b78e6', '#ffffff'];

  function resize() {
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    width = canvas.clientWidth;
    height = canvas.clientHeight;
    canvas.width = width * dpr;
    canvas.height = height * dpr;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  }

  function spawn() {
    return {
      x: Math.random() * width,
      y: Math.random() * height,
      r: 1 + Math.random() * 2.5,
      vx: (Math.random() - 0.5) * 0.25,
      vy: -0.1 - Math.random() * 0.35,
      alpha: 0.15 + Math.random() * 0.55,
      phase: Math.random() * Math.PI * 2,
      color: palette[Math.floor(Math.random() * palette.length)]
    };
  }

  function step(t) {
    ctx.clearRect(0, 0, width, height);
    for (const p of particles) {
      p.x += p.vx + Math.sin(t / 3000 + p.phase) * 0.08;
      p.y += p.vy;
      if (p.y < -10) Object.assign(p, spawn(), { y: height + 10 });
      if (p.x < -10) p.x = width + 10;
      if (p.x > width + 10) p.x = -10;
      const twinkle = 0.6 + 0.4 * Math.sin(t / 900 + p.phase);
      ctx.globalAlpha = p.alpha * twinkle;
      ctx.fillStyle = p.color;
      ctx.beginPath();
      ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
      ctx.fill();
    }
    ctx.globalAlpha = 1;
    raf = requestAnimationFrame(step);
  }

  onMount(() => {
    ctx = canvas.getContext('2d');
    resize();
    particles = Array.from({ length: COUNT }, spawn);
    window.addEventListener('resize', resize);
    raf = requestAnimationFrame(step);
  });

  onDestroy(() => {
    cancelAnimationFrame(raf);
    window.removeEventListener('resize', resize);
  });
</script>

<canvas class="particles" bind:this={canvas} aria-hidden="true"></canvas>

<style>
  .particles {
    position: fixed;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    z-index: 0;
  }
</style>
