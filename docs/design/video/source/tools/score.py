"""Original score for the Ech0 film — code-synthesised, no samples or third-party audio.
96 BPM (beat 0.625s, bar 2.5s), F major, 97.5s. Arrangement follows plan.json shot boundaries:
  open 0–5      soft pad + one piano note on beat 1 (the dot), sparse notes
  echo 5–8.75   a 3-note motif with decaying delay repeats — a literal echo (the product's name)
  write 8.75–22.5  felt-piano ostinato + soft shaker enter (attach, upload, type), room for SFX
  land 22.5–28.75  bass enters on the downbeat: arrival
  detail 28.75–41.25 intimate, lighter pulse under share / like / reply (no kick)
  stream 41.25–48.75 flowing piano melody over the ostinato
  reach 48.75–56.25 fuller voicing, bell partials double the melody
  copilot 56.25–66.25 arpeggiated pulse, melody steps back for reading
  panel 66.25–83.75 the admin tour: steady bright pulse, light melody, room for click SFX
  own 83.75–92.5   warm build (Bb → C lift) under the export page and the docker command
  close 92.5–97.5 resolve to Fmaj9, final chord on beat 152 with a long tail
Usage: python3 tools/score.py assets/music.wav
"""
import sys, wave
import numpy as np
from scipy.signal import fftconvolve, butter, sosfilt

SR = 48000; DUR = 97.5; N = int(SR * DUR)
BEAT = 60 / 96; BAR = 4 * BEAT
rng = np.random.default_rng(1918)
L = np.zeros(N); R = np.zeros(N)
hz = lambda m: 440.0 * 2 ** ((m - 69) / 12)

def add(sig, t0, pan=0.0, gain=1.0):
    i = int(round(t0 * SR));
    if i >= N: return
    sig = sig[: N - i] * gain
    L[i:i + len(sig)] += sig * np.sqrt((1 - pan) / 2)
    R[i:i + len(sig)] += sig * np.sqrt((1 + pan) / 2)

def piano(note, dur, vel=0.5, bright=1.0):
    """Felt-piano-ish: slightly inharmonic partials, fast attack, two-stage decay, hammer noise."""
    n = int((dur + 1.6) * SR); t = np.arange(n) / SR; f = hz(note)
    out = np.zeros(n)
    for k, a in enumerate([1.0, 0.42, 0.22, 0.12, 0.07, 0.04], start=1):
        fk = f * k * np.sqrt(1 + 0.0004 * k * k)
        dec = np.exp(-t * (1.1 + 0.9 * k) * (0.7 + note / 120))
        out += a * (bright if k > 2 else 1) * np.sin(2 * np.pi * fk * t + k) * dec
    env = np.minimum(1, t / 0.004) * (0.55 * np.exp(-t * 2.2) + 0.45 * np.exp(-t * 0.55))
    rel = np.clip((dur + 0.25 - t) / 0.25, 0, 1) ** 2
    ham = rng.normal(0, 1, n) * np.exp(-t * 90) * 0.05
    return (out * env * rel + ham) * vel * 0.32

def pad(notes, dur, vel=0.12, attack=0.8):
    n = int((dur + 1.5) * SR); t = np.arange(n) / SR; out = np.zeros(n)
    for m in notes:
        for det in (-0.07, 0.0, 0.06):
            f = hz(m + det)
            out += np.sin(2 * np.pi * f * t) + 0.18 * np.sin(4 * np.pi * f * t + 0.3)
    env = np.minimum(1, t / attack) * np.clip((dur + 1.2 - t) / 1.2, 0, 1)
    return out * env * vel / (3 * len(notes))

def bass(note, dur, vel=0.35):
    n = int((dur + 0.4) * SR); t = np.arange(n) / SR; f = hz(note)
    s = np.sin(2 * np.pi * f * t) + 0.22 * np.sin(4 * np.pi * f * t) + 0.06 * np.sin(6 * np.pi * f * t)
    env = np.minimum(1, t / 0.01) * np.exp(-t * 1.4) * np.clip((dur + 0.3 - t) / 0.3, 0, 1)
    return s * env * vel

def bell(note, vel=0.12):
    n = int(2.6 * SR); t = np.arange(n) / SR; f = hz(note)
    s = np.sin(2 * np.pi * f * t) + 0.5 * np.sin(2 * np.pi * f * 2.76 * t) * np.exp(-t * 3) + 0.25 * np.sin(2 * np.pi * f * 5.4 * t) * np.exp(-t * 6)
    return s * np.minimum(1, t / 0.002) * np.exp(-t * 1.6) * vel

def shaker(vel=0.03):
    n = int(0.09 * SR); t = np.arange(n) / SR
    s = sosfilt(butter(2, [5000, 11000], 'bandpass', fs=SR, output='sos'), rng.normal(0, 1, n))
    return s * np.exp(-t * 60) * np.minimum(1, t / 0.004) * vel

def soft_kick(vel=0.5):
    n = int(0.35 * SR); t = np.arange(n) / SR
    ph = 2 * np.pi * (46 * t + 40 * 0.03 * (1 - np.exp(-t / 0.03)))
    return np.sin(ph) * np.exp(-t * 11) * np.minimum(1, t / 0.002) * vel

def sub_hit(vel=0.6):
    n = int(2.2 * SR); t = np.arange(n) / SR
    ph = 2 * np.pi * (38 * t + 30 * 0.08 * (1 - np.exp(-t / 0.08)))
    return np.sin(ph) * np.exp(-t * 2.0) * np.minimum(1, t / 0.003) * vel

b = lambda k: k * BEAT          # beat → seconds
# Harmony (F major): Fmaj9, Dm9, Bbmaj9, C6/9 — one chord per bar.
CH = {'F': [53, 57, 60, 64, 67], 'Dm': [50, 53, 57, 60, 64], 'Bb': [46, 50, 53, 57, 60], 'C': [48, 52, 55, 57, 62]}
PROG = ['F', 'Dm', 'Bb', 'C']
chord_at = lambda bar: PROG[bar % 4]

# ——— open (bars 0–1): pad + the dot's note on beat 1 + sparse replies
add(pad([57, 60, 64, 67], 5.0, 0.10, attack=1.6), 0.0)
add(piano(81, 1.5, 0.45), b(1), 0.1)
add(piano(76, 1.2, 0.25), b(4), -0.2)
add(piano(72, 1.4, 0.22), b(6), 0.2)
# ——— echo (5–8.75): motif C6–A5–F5 with three decaying delay repeats (dotted-eighth feel)
motif = [(0, 84), (0.5, 81), (1.0, 77)]
for rep, (g, pan, bright) in enumerate([(0.42, 0.0, 1.0), (0.24, -0.45, 0.8), (0.14, 0.45, 0.6), (0.08, 0.0, 0.45)]):
    for off, n in motif:
        add(piano(n, 0.6, g, bright), 5.0 + rep * b(1.5) + b(off), pan)
add(pad(CH['Dm'][1:], 3.75, 0.10, attack=1.0), 5.0)

# ——— bars from write onward: 8.75 = beat 14 (half-bar); use beat-indexed loops
def ostinato(start, end, vel, pattern=(0, 2, 1, 3, 2, 4, 1, 3)):
    k0 = int(round(start / BEAT * 2)); k1 = int(round(end / BEAT * 2))
    for k in range(k0, k1):
        t = k * BEAT / 2; bar = int(t // BAR); ch = CH[chord_at(bar)]
        note = ch[1:][pattern[k % 8] % 4] + 12
        add(piano(note, 0.35, vel * (1.0 if k % 2 == 0 else 0.7), 0.8), t, -0.25 if k % 2 == 0 else 0.25)

def pads(start, end, vel):
    bar0 = int(start // BAR); bar1 = int(np.ceil(end / BAR))
    for bar in range(bar0, bar1):
        t0 = max(start, bar * BAR); t1 = min(end, (bar + 1) * BAR)
        add(pad(CH[chord_at(bar)][1:], t1 - t0, vel, attack=0.5), t0)

def basses(start, end, vel=0.085):
    bar0 = int(start // BAR); bar1 = int(np.ceil(end / BAR))
    for bar in range(bar0, bar1):
        root = CH[chord_at(bar)][0] - 12
        for off, n in [(0, root), (b(2.5), root), (b(3), root + 7)]:
            t = bar * BAR + off
            if start <= t < end: add(bass(n, 0.55, vel), t)

def shakers(start, end, vel):
    k0 = int(round(start / BEAT * 2)); k1 = int(round(end / BEAT * 2))
    for k in range(k0, k1): add(shaker(vel * (1.0 if k % 2 else 0.55)), k * BEAT / 2, 0.35 if k % 2 else -0.35)

def kicks(start, end, vel=0.28, every=2):
    k0 = int(round(start / BEAT)); k1 = int(round(end / BEAT))
    for k in range(k0, k1):
        if k % every == 0: add(soft_kick(vel), k * BEAT)

# write
pads(8.75, 22.5, 0.09); ostinato(8.75, 22.5, 0.24); shakers(10.0, 22.5, 0.022)
add(bell(84, 0.06), 14.5, 0.2)   # upload finished
# land — bass + kick on the downbeat at 22.5
pads(22.5, 28.75, 0.085); ostinato(22.5, 28.75, 0.22); basses(22.5, 28.75); shakers(22.5, 28.75, 0.026); kicks(22.5, 28.75, 0.06)
add(bell(88, 0.10), 23.125, 0.2)
# detail — Pass it on: intimate, lighter pulse (no kick) under share, like and reply
pads(28.75, 41.25, 0.075); ostinato(28.75, 41.25, 0.15, pattern=(0, 2, 1, 2, 3, 2, 1, 2)); basses(28.75, 41.25, 0.06); shakers(30, 41.25, 0.018)
for off, n, d in [(0, 77, 1.5), (1.5, 79, 0.5), (2, 81, 2), (6, 79, 1), (7, 77, 1), (8, 76, 2), (12, 77, 1.5), (13.5, 79, 0.5), (14, 81, 2)]:
    add(piano(n, d * BEAT, 0.24), 30 + b(off), -0.05)
add(bell(84, 0.06), 32.5, 0.25)
# stream — melody over the ostinato
pads(41.25, 48.75, 0.08); ostinato(41.25, 48.75, 0.18); basses(41.25, 48.75); shakers(41.25, 48.75, 0.026); kicks(41.25, 48.75, 0.06)
mel = [(0, 81, 1.5), (1.5, 79, 0.5), (2, 77, 1), (3, 76, 1), (4, 74, 1.5), (5.5, 76, 0.5), (6, 77, 2), (8, 79, 1.5), (9.5, 81, 0.5), (10, 84, 2)]
for off, n, d in mel: add(piano(n, d * BEAT, 0.34), 41.25 + b(off), 0.05)
# reach — fuller; bells double the line
pads(48.75, 56.25, 0.09); ostinato(48.75, 56.25, 0.20); basses(48.75, 56.25, 0.09); shakers(48.75, 56.25, 0.03); kicks(48.75, 56.25, 0.065)
for off, n in [(0, 84), (2, 81), (4, 79), (6, 77), (8, 81)]: add(bell(n, 0.07), 48.75 + b(off), 0.3)
# copilot — arpeggio pulse, melody backs off for reading
pads(56.25, 66.25, 0.075); ostinato(56.25, 66.25, 0.15, pattern=(0, 1, 2, 3, 2, 1, 3, 2)); basses(56.25, 66.25, 0.075); shakers(56.25, 66.25, 0.02); kicks(56.25, 66.25, 0.05)
# panel — the admin tour: steady, bright, uncluttered (clicks carry the rhythm on screen)
pads(66.25, 83.75, 0.08); ostinato(66.25, 83.75, 0.19, pattern=(0, 2, 3, 1, 2, 4, 3, 1)); basses(66.25, 83.75, 0.085); shakers(66.25, 83.75, 0.028); kicks(66.25, 83.75, 0.06)
tour = [(0, 81, 2), (2, 79, 1), (3, 77, 1), (4, 76, 2), (6, 77, 2), (8, 79, 1.5), (9.5, 81, 0.5), (10, 84, 2), (12, 81, 2), (14, 79, 2)]
for off, n, d in tour: add(piano(n, d * BEAT, 0.26), 68.75 + b(off), 0.1)
for off, n in [(0, 88), (8, 86), (16, 84)]: add(bell(n, 0.05), 71.25 + b(off), -0.3)
# own — warm build: Bb → C lift over two bars, pulse thins so the command reads
pads(83.75, 92.5, 0.09); ostinato(83.75, 90, 0.16); basses(83.75, 90, 0.08); shakers(83.75, 92.5, 0.02)
for k, n in enumerate([74, 76, 77, 79, 81, 79, 77, 81, 84]): add(piano(n, 0.9, 0.2 + 0.015 * k), 85 + k * b(1.5), 0.05)
add(pad(CH['C'][1:], 2.5, 0.07, attack=1.2), 90)
# close — resolve: Bb → C lift, then Fmaj9 on beat 152 with a long tail
add(pad(CH['Bb'][1:], 2.5, 0.09, attack=0.3), 92.5)
add(piano(77, 0.8, 0.28), 92.5 + b(0)); add(piano(79, 0.8, 0.26), 92.5 + b(1)); add(piano(81, 1.0, 0.3), 92.5 + b(2))
final = 95
for i, n in enumerate([41, 53, 57, 60, 64, 67, 72]): add(piano(n, 2.4, 0.30 if n > 50 else 0.2), final + i * 0.012, (i - 3) * 0.12)
add(pad([57, 60, 64, 67], 2.5, 0.08, attack=0.2), final)
add(bell(84, 0.08), final, -0.2)

# Room: short synthetic hall, then a gentle bus.
ir_n = int(2.0 * SR); ti = np.arange(ir_n) / SR
ir = rng.normal(0, 1, ir_n) * np.exp(-ti * 3.2); ir[0] = 0; ir /= np.sqrt((ir ** 2).sum())
wetL = fftconvolve(L, ir)[:N]; wetR = fftconvolve(R, np.roll(ir, 97))[:N]
L2 = L * 0.82 + wetL * 0.30; R2 = R * 0.82 + wetR * 0.30
hp = butter(2, 45, 'highpass', fs=SR, output='sos'); L2 = sosfilt(hp, L2); R2 = sosfilt(hp, R2)
# Low shelf ≈ -8 dB under 120 Hz: synthesised lows read heavier than they sound on small speakers.
lp = butter(2, 120, 'lowpass', fs=SR, output='sos'); L2 = L2 - 0.6 * sosfilt(lp, L2); R2 = R2 - 0.6 * sosfilt(lp, R2)
# Gentle air: +~5 dB high shelf above 2.5 kHz so the felt piano still speaks next to UI sounds.
hs = butter(2, 2500, 'highpass', fs=SR, output='sos'); L2 = L2 + 0.8 * sosfilt(hs, L2); R2 = R2 + 0.8 * sosfilt(hs, R2)
fade = np.clip((DUR - np.arange(N) / SR) / 1.2, 0, 1); fade_in = np.minimum(1, np.arange(N) / (0.05 * SR))
L2 *= fade * fade_in; R2 *= fade * fade_in
peak = max(np.abs(L2).max(), np.abs(R2).max()); g = 10 ** (-1.5 / 20) / peak
out = np.stack([L2 * g, R2 * g], axis=1)
with wave.open(sys.argv[1] if len(sys.argv) > 1 else 'assets/music.wav', 'wb') as w:
    w.setnchannels(2); w.setsampwidth(2); w.setframerate(SR)
    w.writeframes((np.clip(out, -1, 1) * 32767).astype('<i2').tobytes())
print('wrote score', DUR, 's')
