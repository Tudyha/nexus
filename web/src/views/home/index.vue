<template>
  <div>
    <!-- Hero Section -->
    <section class="hero min-h-[80vh] bg-base-200">
      <div class="hero-content text-center">
        <div class="max-w-3xl">
          <h1
            class="text-4xl md:text-6xl font-bold mb-6 bg-linear-to-r from-primary to-secondary bg-clip-text text-transparent">
            {{ t('home.hero.title') }}
          </h1>
          <p class="text-lg md:text-xl mb-8 opacity-80">
            {{ t('home.hero.subtitle') }}
          </p>
          <div class="flex flex-col sm:flex-row gap-4 justify-center">
            <router-link :to="{ name: 'Dashboard' }" class="btn btn-primary btn-lg">
              {{ t('home.hero.cta') }}
            </router-link>
            <a href="#features" class="btn btn-outline btn-lg">
              {{ t('home.hero.learnMore') }}
            </a>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section id="features" class="py-20 px-4 bg-base-100">
      <div class="max-w-6xl mx-auto">
        <div class="text-center mb-16">
          <h2 class="text-3xl md:text-4xl font-bold mb-4">{{ t('home.features.title') }}</h2>
          <p class="text-base-content/60 text-lg">{{ t('home.features.subtitle') }}</p>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          <div v-for="feature in features" :key="feature.key" class="card bg-base-200 border border-base-300 shadow-sm hover:shadow-md transition-shadow">
            <div class="card-body items-center text-center">
              <Icon :icon="feature.icon" class="w-10 h-10 text-primary mb-2" />
              <h3 class="card-title text-lg">{{ t(feature.titleKey) }}</h3>
              <p class="text-base-content/60 text-sm">{{ t(feature.descKey) }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- How It Works Section -->
    <section class="py-20 px-4 bg-base-200">
      <div class="max-w-5xl mx-auto">
        <div class="text-center mb-16">
          <h2 class="text-3xl md:text-4xl font-bold mb-4">{{ t('home.howItWorks.title') }}</h2>
          <p class="text-base-content/60 text-lg">{{ t('home.howItWorks.subtitle') }}</p>
        </div>
        <div class="flex flex-col md:flex-row gap-8 justify-center items-start">
          <div v-for="(step, index) in steps" :key="index" class="flex-1 text-center">
            <div class="w-16 h-16 rounded-full bg-primary text-primary-content text-2xl font-bold flex items-center justify-center mx-auto mb-4">
              {{ index + 1 }}
            </div>
            <h3 class="text-xl font-bold mb-2">{{ t(step.titleKey) }}</h3>
            <p class="text-base-content/60">{{ t(step.descKey) }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="py-20 px-4 bg-linear-to-r from-primary to-secondary text-primary-content">
      <div class="max-w-3xl mx-auto text-center">
        <h2 class="text-3xl md:text-4xl font-bold mb-4">{{ t('home.cta.title') }}</h2>
        <p class="text-lg mb-8 opacity-90">{{ t('home.cta.subtitle') }}</p>
        <router-link :to="{ name: 'Dashboard' }" class="btn btn-lg bg-white text-primary hover:bg-gray-100 border-0">
          {{ t('home.cta.button') }}
        </router-link>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'

const { t } = useI18n()

const features = [
  { key: 'nat', icon: 'material-symbols:lan-outline', titleKey: 'home.features.nat.title', descKey: 'home.features.nat.desc' },
  { key: 'secure', icon: 'material-symbols:lock-outline', titleKey: 'home.features.secure.title', descKey: 'home.features.secure.desc' },
  { key: 'manage', icon: 'material-symbols:web-dashboard', titleKey: 'home.features.manage.title', descKey: 'home.features.manage.desc' },
  { key: 'multi', icon: 'material-symbols:alt-route', titleKey: 'home.features.multi.title', descKey: 'home.features.multi.desc' },
  { key: 'monitor', icon: 'material-symbols:monitoring', titleKey: 'home.features.monitor.title', descKey: 'home.features.monitor.desc' },
  { key: 'cross', icon: 'material-symbols:devices-outline', titleKey: 'home.features.cross.title', descKey: 'home.features.cross.desc' },
]

const steps = [
  { titleKey: 'home.howItWorks.step1.title', descKey: 'home.howItWorks.step1.desc' },
  { titleKey: 'home.howItWorks.step2.title', descKey: 'home.howItWorks.step2.desc' },
  { titleKey: 'home.howItWorks.step3.title', descKey: 'home.howItWorks.step3.desc' },
]
</script>
