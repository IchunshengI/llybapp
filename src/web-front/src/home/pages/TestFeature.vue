<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { REGIONS_CN_MINI } from "../data/regions-cn-mini.js";

const form = reactive({
  gender: "male",
  solarDate: "",
  birthTime: "",
  provinceCode: "",
  cityCode: "",
});

const genderText = computed(() => (form.gender === "male" ? "男" : "女"));

const apiBase = (() => {
  const v = typeof __BACKEND_HOST__ === "string" ? __BACKEND_HOST__.trim() : "";
  return v ? v.replace(/\/+$/, "") : "";
})();

const reasoningLoading = ref(false);
const reasoningError = ref("");
const reasoningResp = ref(null); // { code, message, result_json }

const prettyResultJson = computed(() => {
  const raw = reasoningResp.value?.result_json;
  if (!raw) return "";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return String(raw);
  }
});

const regionsLoading = ref(true);
const regionsError = ref("");
const regions = ref([]);

const REGIONS_CACHE_KEY = "llyb.regions.pca-code.v1";
const REGIONS_URLS = [
  // "Realtime from internet" (may change over time; keep fallbacks).
  "https://unpkg.com/china-division@latest/dist/pca-code.json",
  // Pinned fallback in case latest is temporarily broken.
  "https://unpkg.com/china-division@2.4.0/dist/pca-code.json",
];

const loadRegions = async () => {
  regionsLoading.value = true;
  regionsError.value = "";

  // 1) Use cached data first (fast path)
  try {
    const cached = JSON.parse(localStorage.getItem(REGIONS_CACHE_KEY) || "null");
    if (cached && Array.isArray(cached.data) && cached.data.length) {
      regions.value = cached.data;
      regionsLoading.value = false;
      // Continue to refresh in background.
    }
  } catch {
    // ignore cache parse errors
  }

  // 2) Fetch fresh data
  for (const url of REGIONS_URLS) {
    try {
      const res = await fetch(url, { cache: "no-store" });
      if (!res.ok) throw new Error(`regions http ${res.status}`);
      const data = await res.json();
      if (!Array.isArray(data) || !data.length) throw new Error("regions invalid");

      regions.value = data;
      try {
        localStorage.setItem(
          REGIONS_CACHE_KEY,
          JSON.stringify({ ts: Date.now(), data })
        );
      } catch {
        // ignore cache quota errors
      }
      regionsLoading.value = false;
      return;
    } catch (e) {
      regionsError.value = String(e?.message || e);
    }
  }

  // 3) Fallback to built-in minimal list
  regions.value = REGIONS_CN_MINI;
  regionsLoading.value = false;
};

onMounted(() => {
  loadRegions();
});

const provinces = computed(() => regions.value);
const selectedProvince = computed(
  () => provinces.value.find((p) => p.code === form.provinceCode) || null
);
const cities = computed(() => selectedProvince.value?.children ?? []);
const selectedCity = computed(
  () => cities.value.find((c) => c.code === form.cityCode) || null
);

const provinceName = computed(() => selectedProvince.value?.name ?? "");
const cityName = computed(() => selectedCity.value?.name ?? "");

const regionText = computed(() =>
  [provinceName.value, cityName.value]
    .filter(Boolean)
    .join(" / ")
);

const handleProvinceChange = () => {
  form.cityCode = "";
};

const dateEl = ref(null);
const openDatePicker = () => {
  const el = dateEl.value;
  if (!el) return;
  // Chrome/Edge support showPicker(); other browsers will just focus.
  if (typeof el.showPicker === "function") el.showPicker();
  else el.focus();
};

const timeEl = ref(null);
const openTimePicker = () => {
  const el = timeEl.value;
  if (!el) return;
  // Chrome/Edge support showPicker(); other browsers will just focus.
  if (typeof el.showPicker === "function") el.showPicker();
  else el.focus();
};

const handleReasoning = async () => {
  if (reasoningLoading.value) return;
  reasoningError.value = "";

  if (!apiBase) {
    reasoningError.value = "未配置 BACKEND_HOST，无法请求后端";
    return;
  }

  reasoningLoading.value = true;
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 10_000);
  try {
    const payload = {
      gender: form.gender === "male" ? 1 : 2, // proto enum: 1=male, 2=female
      solar_date: form.solarDate,
      birth_time: form.birthTime,
      province: provinceName.value,
      city: cityName.value,
    };

    console.log("reasoning ->", `${apiBase}/admin/reasoning`, payload);
    const res = await fetch(`${apiBase}/admin/reasoning`, {
      method: "POST",
      headers: { "Content-Type": "text/plain" },
      body: JSON.stringify(payload),
      signal: controller.signal,
    });
    const raw = await res.text();
    if (!res.ok) {
      // Backend may return plain text error for 500s; surface it to help debugging.
      throw new Error(`reasoning http ${res.status}: ${raw.slice(0, 500)}`);
    }
    let data = {};
    try {
      data = raw ? JSON.parse(raw) : {};
    } catch {
      data = { code: -1, message: "后端返回非 JSON", result_json: raw };
    }
    reasoningResp.value = data;
  } catch (e) {
    reasoningError.value = `请求失败：${e?.message || "unknown error"}`;
  } finally {
    clearTimeout(timer);
    reasoningLoading.value = false;
  }
};
</script>

<template>
  <div class="layout">
    <section class="card">
      <h2>基础推理</h2>

      <form class="form" @submit.prevent>
        <div class="row">
          <div class="field">
            <span>性别</span>
            <div class="seg">
              <label class="seg-item">
                <input v-model="form.gender" type="radio" value="male" />
                <span>男</span>
              </label>
              <label class="seg-item">
                <input v-model="form.gender" type="radio" value="female" />
                <span>女</span>
              </label>
            </div>
          </div>

          <label class="field">
            <span>出生日期(公历/阳历)</span>
            <div class="datewrap" @click="openDatePicker">
              <input
                ref="dateEl"
                v-model="form.solarDate"
                type="date"
                @click.stop="openDatePicker"
              />
            </div>
          </label>
        </div>

        <label class="field">
          <span>出生时间</span>
          <div class="datewrap" @click="openTimePicker">
            <input
              ref="timeEl"
              v-model="form.birthTime"
              type="time"
              placeholder="例如：08:30"
              @click.stop="openTimePicker"
            />
          </div>
        </label>

        <div class="field">
          <span>地区(省 / 市)</span>
          <div class="region">
            <select
              v-model="form.provinceCode"
              :disabled="regionsLoading"
              @change="handleProvinceChange"
            >
              <option value="" disabled>
                {{ regionsLoading ? "地区数据加载中..." : "请选择省" }}
              </option>
              <option v-for="p in provinces" :key="p.code" :value="p.code">
                {{ p.name }}
              </option>
            </select>
            <select
              v-model="form.cityCode"
              :disabled="!form.provinceCode || regionsLoading"
            >
              <option value="" disabled>请选择市</option>
              <option v-for="c in cities" :key="c.code" :value="c.code">
                {{ c.name }}
              </option>
            </select>
          </div>
          <div v-if="regionsError" class="hint">
            地区数据拉取失败，已使用内置数据（可点这里重试：
            <button class="linkish" type="button" @click="loadRegions">重试</button>
            ）
          </div>
        </div>

        <button class="primary" type="button" :disabled="reasoningLoading" @click="handleReasoning">
          {{ reasoningLoading ? "排盘中..." : "排盘" }}
        </button>
        <div v-if="reasoningError" class="hint">{{ reasoningError }}</div>
      </form>
    </section>

    <section class="card preview">
      <h3>当前输入</h3>
      <div class="kv">
        <div class="k">性别</div>
        <div class="v">{{ genderText }}</div>
      </div>
      <div class="kv">
        <div class="k">出生日期</div>
        <div class="v">{{ form.solarDate || "未选择" }}</div>
      </div>
      <div class="kv">
        <div class="k">出生时间</div>
        <div class="v">{{ form.birthTime || "未选择" }}</div>
      </div>
      <div class="kv">
        <div class="k">地区</div>
        <div class="v">{{ regionText || "未填写" }}</div>
      </div>

      <div class="tip">排盘结果</div>
      <div v-if="reasoningResp" class="result">
        <div class="kv">
          <div class="k">code</div>
          <div class="v">{{ reasoningResp.code }}</div>
        </div>
        <div class="kv">
          <div class="k">message</div>
          <div class="v">{{ reasoningResp.message || "-" }}</div>
        </div>
        <div class="kv">
          <div class="k">result</div>
          <div class="v">
            <pre class="pre">{{ prettyResultJson || "-" }}</pre>
          </div>
        </div>
      </div>
      <div v-else class="tip">点击“排盘”后，这里展示后端返回结果。</div>
    </section>
  </div>
</template>

<style scoped>
.layout {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(0, 0.8fr);
  gap: 16px;
}

.card {
  border-radius: 18px;
  padding: 18px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.10);
}

.field {
  color: rgba(255, 255, 255, 0.78);
}

.row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 14px;
}

.seg {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.seg-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.10);
  background: rgba(0, 0, 0, 0.18);
  color: rgba(255, 255, 255, 0.88);
}

.seg-item input {
  accent-color: rgba(77, 215, 200, 1);
}

.region {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.region select,
.field input[type="date"],
.field input[type="time"],
.field input[type="text"] {
  background: rgba(0, 0, 0, 0.18);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.92);
}

.region select:focus,
.field input[type="date"]:focus,
.field input[type="time"]:focus,
.field input[type="text"]:focus {
  border-color: rgba(77, 215, 200, 0.7);
  box-shadow: 0 0 0 3px rgba(77, 215, 200, 0.18);
}

.region select {
  padding: 12px 14px;
  border-radius: 12px;
  font-size: 1rem;
  appearance: none;
  cursor: pointer;
}

.field input[type="date"],
.field input[type="time"],
.field input[type="text"] {
  padding: 12px 14px;
  border-radius: 12px;
  font-size: 1rem;
  width: 100%;
  box-sizing: border-box;
}

.region select:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.datewrap {
  cursor: pointer;
}

.datewrap input[type="date"] {
  width: 100%;
  cursor: pointer;
}

.datewrap input[type="time"] {
  width: 100%;
  cursor: pointer;
}

.hint {
  margin-top: 10px;
  font-size: 12px;
  opacity: 0.7;
  line-height: 1.5;
}

.linkish {
  background: none;
  border: none;
  color: rgba(77, 215, 200, 0.95);
  font-weight: 700;
  padding: 0;
  cursor: pointer;
}

.linkish:hover {
  text-decoration: underline;
}

.preview h3 {
  font-size: 14px;
  margin-bottom: 12px;
  opacity: 0.85;
}

.kv {
  display: grid;
  grid-template-columns: 110px minmax(0, 1fr);
  gap: 10px;
  padding: 10px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.k {
  opacity: 0.7;
  font-size: 12px;
}

.v {
  font-size: 13px;
  word-break: break-word;
}

.tip {
  margin-top: 14px;
  font-size: 12px;
  opacity: 0.65;
  line-height: 1.6;
}

.result {
  margin-top: 6px;
}

.pre {
  margin: 0;
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.10);
  background: rgba(0, 0, 0, 0.18);
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
  line-height: 1.5;
}

@media (max-width: 900px) {
  .layout {
    grid-template-columns: 1fr;
  }

  .row {
    grid-template-columns: 1fr;
  }
}
</style>
