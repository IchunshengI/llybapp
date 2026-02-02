<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { REGIONS_CN_MINI } from "../data/regions-cn-mini.js";

const form = reactive({
  bazi: "",
  gender: "male",
  solarDate: "",
  provinceCode: "",
  cityCode: "",
  districtCode: "",
});

const genderText = computed(() => (form.gender === "male" ? "男" : "女"));

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
const districts = computed(() => selectedCity.value?.children ?? []);

const provinceName = computed(() => selectedProvince.value?.name ?? "");
const cityName = computed(() => selectedCity.value?.name ?? "");
const districtName = computed(
  () => districts.value.find((d) => d.code === form.districtCode)?.name ?? ""
);

const regionText = computed(() =>
  [provinceName.value, cityName.value, districtName.value]
    .filter(Boolean)
    .join(" / ")
);

const handleProvinceChange = () => {
  form.cityCode = "";
  form.districtCode = "";
};

const handleCityChange = () => {
  form.districtCode = "";
};

const dateEl = ref(null);
const openDatePicker = () => {
  const el = dateEl.value;
  if (!el) return;
  // Chrome/Edge support showPicker(); other browsers will just focus.
  if (typeof el.showPicker === "function") el.showPicker();
  else el.focus();
};
</script>

<template>
  <div class="layout">
    <section class="card">
      <h2>基础推理</h2>

      <form class="form" @submit.prevent>
        <label class="field">
          <span>生辰八字</span>
          <input v-model="form.bazi" type="text" placeholder="例如：甲子 乙丑 丙寅 丁卯" />
        </label>

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
            <span>公历(阳历)</span>
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

        <div class="field">
          <span>地区(省 / 市 / 区)</span>
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
              @change="handleCityChange"
            >
              <option value="" disabled>请选择市</option>
              <option v-for="c in cities" :key="c.code" :value="c.code">
                {{ c.name }}
              </option>
            </select>
            <select
              v-model="form.districtCode"
              :disabled="!form.cityCode || regionsLoading"
            >
              <option value="" disabled>请选择区</option>
              <option v-for="d in districts" :key="d.code" :value="d.code">
                {{ d.name }}
              </option>
            </select>
          </div>
          <div v-if="regionsError" class="hint">
            地区数据拉取失败，已使用内置数据（可点这里重试：
            <button class="linkish" type="button" @click="loadRegions">重试</button>
            ）
          </div>
        </div>

        <button class="primary" type="button">开始测算（占位）</button>
      </form>
    </section>

    <section class="card preview">
      <h3>当前输入</h3>
      <div class="kv">
        <div class="k">生辰八字</div>
        <div class="v">{{ form.bazi || "未填写" }}</div>
      </div>
      <div class="kv">
        <div class="k">性别</div>
        <div class="v">{{ genderText }}</div>
      </div>
      <div class="kv">
        <div class="k">公历(阳历)</div>
        <div class="v">{{ form.solarDate || "未选择" }}</div>
      </div>
      <div class="kv">
        <div class="k">地区</div>
        <div class="v">{{ regionText || "未填写" }}</div>
      </div>

      <div class="tip">后端就绪后，这里可以展示接口返回的测算结果。</div>
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
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.region select,
.field input[type="date"],
.field input[type="text"] {
  background: rgba(0, 0, 0, 0.18);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.92);
}

.region select:focus,
.field input[type="date"]:focus,
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

@media (max-width: 900px) {
  .layout {
    grid-template-columns: 1fr;
  }

  .row {
    grid-template-columns: 1fr;
  }
}
</style>
