import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

const resources = {
  en: {
    translation: {
      status: {
        preparing: 'Preparing Desktop',
        spinning_up: 'We are spinning up your personal environment.',
        connecting: 'Connecting to {{title}}...',
        authenticating: 'Authenticating...',
        label: 'Status:',
        unknown: 'unknown',
      },
      error: {
        title: 'Error Occurred',
        generic_desc: 'Something went wrong while connecting.',
        connection_failed: 'Failed to connect to server',
        login_failed: 'Login failed',
      },
      action: {
        retry: 'Retry',
      },
      login: {
        oidc_button: 'Login with {{name}}',
        or_continue: 'Or continue with',
        anonymous_button: 'Continue as Guest',
      },
      legal: {
        agreement:
          'By continuing, you agree to our <0>Terms of Service</0> and <1>Privacy Policy</1>.',
        agreement_tos_only:
          'By continuing, you agree to our <0>Terms of Service</0>.',
        agreement_privacy_only:
          'By continuing, you agree to our <0>Privacy Policy</0>.',
      },
    },
  },
  ja: {
    translation: {
      status: {
        preparing: 'デスクトップを準備中',
        spinning_up: 'パーソナル環境を起動しています。',
        connecting: '{{title}}へ接続中...',
        authenticating: '認証中...',
        label: 'ステータス:',
        unknown: '不明',
      },
      error: {
        title: 'エラーが発生しました',
        generic_desc: '接続中に問題が発生しました。',
        connection_failed: 'サーバーへの接続に失敗しました',
        login_failed: 'ログインに失敗しました',
      },
      action: {
        retry: '再試行',
      },
      login: {
        oidc_button: '{{name}}でログイン',
        or_continue: 'または',
        anonymous_button: 'ゲストでログイン',
      },
      legal: {
        agreement:
          '続行することで、<0>利用規約</0>および<1>プライバシーポリシー</1>に同意したものとみなされます。',
        agreement_tos_only:
          '続行することで、<0>利用規約</0>に同意したものとみなされます。',
        agreement_privacy_only:
          '続行することで、<0>プライバシーポリシー</0>に同意したものとみなされます。',
      },
    },
  },
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    interpolation: {
      escapeValue: false, // react already safes from xss
    },
  });

export default i18n;
