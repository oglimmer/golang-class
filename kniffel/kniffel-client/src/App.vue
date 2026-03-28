<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Ref } from 'vue'
import createClient from 'openapi-fetch';
import type { components, paths } from '@/api/v1';

let client = createClient<paths>();

interface PlayerInformation {
    index: number;
    name: string;
}

const names : Ref<PlayerInformation[]> = ref([{index: 0, name: ''}, {index: 1, name: ''}]);
const gameData : Ref<components["schemas"]["GameResponse"]|undefined> = ref();
const rerollSelection = ref([false, false, false, false, false]);
// the booking type selected in the dropdown box
const selectedBookingType : Ref<components["schemas"]["GameResponse"]["usedBookingTypes"]|undefined> = ref();
const apiServer = ref(`${__API_URL__}`);

// map dice value (1-6) to a Unicode die face
const diceFaces = ['', '\u2680', '\u2681', '\u2682', '\u2683', '\u2684', '\u2685'];

// the game is over when there are no more available booking types
const gameOver = computed(() =>
    gameData.value?.gameId && gameData.value?.availableBookingTypes?.length === 0
);

async function createGame() {
    client = createClient<paths>({ baseUrl: apiServer.value });
    const { data } = await client.POST("/api/v1/game/", {
        body: {
            playerNames: names.value.map(n => n.name)
        }
    });
    gameData.value = data;
}

async function reroll() {
    const diceToKeep : number[] = [];
    if (gameData.value) {
        for (let i = 0; i < gameData.value.diceRolls.length; i++) {
            if (rerollSelection.value[i]) {
                diceToKeep.push(gameData.value.diceRolls[i]);
            }
        }
        const { data } = await client.POST("/api/v1/game/{gameId}/roll", {
            params: {
                path: {
                    gameId: gameData.value.gameId
                }
            },
            body: {
                diceToKeep
            }
        });
        gameData.value = data;
        selectedBookingType.value = undefined;
        if (gameData.value) {
            for (let i = 0; i < gameData.value.diceRolls.length; i++) {
                const idxToKeep = diceToKeep.indexOf(gameData.value.diceRolls[i]);
                if (idxToKeep === -1) {
                    rerollSelection.value[i] = false;
                } else {
                    rerollSelection.value[i] = true;
                    diceToKeep.splice(idxToKeep, 1);
                }
            }
        }
    }
}

// simple REST API call to send the booking type
async function book() {
  if (gameData.value) {
    const { data } = await client.POST("/api/v1/game/{gameId}/book", {
      params: {
        path: {
          gameId: gameData.value.gameId
        }
      },
      body: {
        bookingType: selectedBookingType.value
      }
    });
    gameData.value = data;
    rerollSelection.value = [false, false, false, false, false];
  }
}

</script>

<template>
  <!-- Setup screen: server selection and player names -->
  <div v-if="!gameData?.gameId" class="card">
    <h1>Kniffel</h1>

    <label class="label">API Server</label>
    <select v-model="apiServer" class="select">
      <option>https://api-kniffel.oglimmer.com</option>
      <option>https://api-rust-kniffel.oglimmer.com</option>
      <option>http://localhost:8080</option>
    </select>

    <label class="label">Players</label>
    <div v-for="ply in names" :key="ply.index" class="player-input">
      <span class="player-number">{{ ply.index + 1 }}</span>
      <input type="text" v-model="ply.name" :placeholder="'Player ' + (ply.index + 1)" class="input" />
    </div>

    <div class="actions">
      <button class="btn btn-secondary" @click="names.push({ index: names.length, name: '' })">
        + Add Player
      </button>
      <button class="btn btn-primary" @click="createGame">Start Game</button>
    </div>
  </div>

  <!-- Game screen -->
  <template v-if="gameData?.gameId">

    <!-- Scoreboard -->
    <div class="card">
      <h2>Scoreboard</h2>
      <div class="scoreboard">
        <div
          v-for="ply in gameData?.playerData"
          :key="ply.name"
          class="score-item"
          :class="{ active: ply.name === gameData.currentPlayerName }"
        >
          <span class="score-name">{{ ply.name }}</span>
          <span class="score-value">{{ ply.score }}</span>
        </div>
      </div>
    </div>

    <!-- Game over -->
    <div v-if="gameOver" class="card card-accent">
      <h2>Game Over!</h2>
    </div>

    <!-- Roll phase: select dice to keep, then re-roll -->
    <div v-if="gameData?.state === 'ROLL'" class="card">
      <div class="roll-header">
        <h2>{{ gameData.currentPlayerName }}'s Turn</h2>
        <span class="badge">Roll {{ gameData?.rollRound }} / 3</span>
      </div>

      <label class="label">Your dice — click to keep:</label>
      <div class="dice-row">
        <button
          v-for="(die, idx) in gameData.diceRolls"
          :key="idx"
          class="die"
          :class="{ kept: rerollSelection[idx] }"
          @click="rerollSelection[idx] = !rerollSelection[idx]"
        >
          {{ diceFaces[die] }}
        </button>
      </div>

      <label class="label">Available categories:</label>
      <div class="categories">
        <span v-for="cat in gameData?.availableBookingTypes" :key="cat" class="category-tag">
          {{ cat }}
        </span>
      </div>

      <button class="btn btn-primary full-width" @click="reroll">Roll Dice</button>
    </div>

    <!-- Book phase: select a category to score -->
    <div v-if="gameData?.state === 'BOOK'" class="card">
      <h2>Book Your Score</h2>

      <div class="dice-row">
        <span v-for="(die, idx) in gameData.diceRolls" :key="idx" class="die">
          {{ diceFaces[die] }}
        </span>
      </div>

      <label class="label">Select category:</label>
      <select v-model="selectedBookingType" class="select">
        <option disabled :value="undefined">Choose...</option>
        <option v-for="cat in gameData.availableBookingTypes" :key="cat" :value="cat">
          {{ cat }}
        </option>
      </select>

      <button class="btn btn-primary full-width" @click="book">Book</button>
    </div>
  </template>
</template>

<style scoped>
/* Card layout */
.card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1.5rem;
  margin-bottom: 1rem;
}

.card-accent {
  border-color: var(--color-accent);
  text-align: center;
}

h1 {
  font-size: 2rem;
  font-weight: 900;
  margin-bottom: 1.5rem;
}

h2 {
  font-size: 1.3rem;
  font-weight: 700;
  margin-bottom: 1rem;
}

/* Labels */
.label {
  display: block;
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--color-text-soft);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 0.5rem;
  margin-top: 1rem;
}

/* Form elements */
.select,
.input {
  width: 100%;
  padding: 0.6rem 0.8rem;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  font-family: inherit;
  font-size: 1rem;
  background: var(--color-bg);
  color: var(--color-text);
}

.select:focus,
.input:focus {
  outline: 2px solid var(--color-accent);
  outline-offset: -1px;
}

/* Player name inputs */
.player-input {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  margin-bottom: 0.5rem;
}

.player-number {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-accent);
  color: white;
  border-radius: 50%;
  font-size: 0.85rem;
  font-weight: 700;
  flex-shrink: 0;
}

/* Buttons */
.btn {
  padding: 0.6rem 1.2rem;
  border: none;
  border-radius: 6px;
  font-family: inherit;
  font-size: 1rem;
  font-weight: 700;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-primary {
  background: var(--color-accent);
  color: white;
}

.btn-primary:hover {
  background: var(--color-accent-hover);
}

.btn-secondary {
  background: var(--color-bg);
  color: var(--color-text);
  border: 1px solid var(--color-border);
}

.btn-secondary:hover {
  border-color: var(--color-text-soft);
}

.full-width {
  width: 100%;
  margin-top: 1rem;
}

.actions {
  display: flex;
  gap: 0.8rem;
  margin-top: 1.5rem;
}

/* Scoreboard */
.scoreboard {
  display: flex;
  gap: 0.8rem;
  flex-wrap: wrap;
}

.score-item {
  flex: 1;
  min-width: 100px;
  padding: 0.8rem;
  background: var(--color-bg);
  border: 2px solid var(--color-border);
  border-radius: 8px;
  text-align: center;
}

.score-item.active {
  border-color: var(--color-accent);
}

.score-name {
  display: block;
  font-size: 0.85rem;
  color: var(--color-text-soft);
}

.score-value {
  display: block;
  font-size: 1.6rem;
  font-weight: 900;
}

/* Roll header with badge */
.roll-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.roll-header h2 {
  margin-bottom: 0;
}

.badge {
  background: var(--color-dice-selected);
  padding: 0.25rem 0.7rem;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 700;
}

/* Dice */
.dice-row {
  display: flex;
  gap: 0.6rem;
  justify-content: center;
  margin: 0.8rem 0;
}

.die {
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2.4rem;
  background: var(--color-dice-bg);
  border: 2px solid var(--color-border);
  border-radius: 10px;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
  /* reset button styles for clickable dice */
  padding: 0;
  font-family: inherit;
  color: var(--color-text);
}

.die:hover {
  border-color: var(--color-accent);
}

.die.kept {
  background: var(--color-dice-selected);
  border-color: var(--color-accent);
}

/* Category tags */
.categories {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-bottom: 0.5rem;
}

.category-tag {
  font-size: 0.8rem;
  padding: 0.2rem 0.6rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 4px;
}
</style>
