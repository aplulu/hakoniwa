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
      dashboard: {
        description: 'Your desktop environment is ready to be created. Click the button below to start.',
        title: 'My Workspaces',
        subtitle: 'Manage your cloud development environments.',
      },
      workspace: {
        status: {
          pending: 'Starting',
          running: 'Running',
          terminating: 'Stopping',
        },
        action: {
          open: 'Open',
          delete: 'Delete Workspace',
          cancel: 'Cancel',
          back_to_list: 'Back to list',
        },
        create: {
          card_title: 'New Workspace',
          card_desc: 'Launch a new environment',
          modal_title: 'Create New Workspace',
          modal_desc: 'Choose a configuration for your new cloud workspace.',
          env_type_label: 'Environment Type',
          submit: 'Launch Workspace',
          placeholder_select: 'Select type',
          no_types: 'No instance types available',
          enable_persistent: 'Enable Persistent Storage',
          persistent_disabled_hint: 'Persistent storage is only available for authenticated users',
          persistent_disabled_global: 'Persistent storage is disabled by the administrator',
          persistent_disabled_type: 'This instance type does not support persistent storage',
        },
      },
      user: {
        guest: 'Guest',
        logout: 'Logout',
      },
      error: {
        title: 'Error Occurred',
        generic_desc: 'Something went wrong while connecting.',
        connection_failed: 'Failed to connect to server',
        login_failed: 'Login failed',
        max_instances: 'Maximum number of instances reached. Please try again later.',
      },
      action: {
        retry: 'Retry',
        start_desktop: 'Start Desktop',
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
      dashboard: {
        description: 'デスクトップ環境を作成する準備ができました。下のボタンをクリックして開始してください。',
        title: 'ワークスペース',
        subtitle: 'クラウド開発環境を管理します。',
      },
      workspace: {
        status: {
          pending: '起動中',
          running: '実行中',
          terminating: '停止中',
        },
        action: {
          open: '開く',
          delete: 'ワークスペースを削除',
          cancel: 'キャンセル',
          back_to_list: '一覧に戻る',
        },
        create: {
          card_title: '新規作成',
          card_desc: '新しい環境を立ち上げます',
          modal_title: 'ワークスペースの作成',
          modal_desc: '新しいクラウドワークスペースの種類を選択してください。',
          env_type_label: '環境タイプ',
          submit: '作成する',
          placeholder_select: 'タイプを選択',
          no_types: '利用可能なタイプがありません',
          enable_persistent: '永続化ボリュームを有効にする',
          persistent_disabled_hint: '永続化ボリュームは認証済みユーザーのみ利用可能です',
          persistent_disabled_global: '永続化ボリュームは管理者によって無効化されています',
          persistent_disabled_type: 'このタイプは永続化ボリュームに対応していません',
        },
      },
      user: {
        guest: 'ゲスト',
        logout: 'ログアウト',
      },
      error: {
        title: 'エラーが発生しました',
        generic_desc: '接続中に問題が発生しました。',
        connection_failed: 'サーバーへの接続に失敗しました',
        login_failed: 'ログインに失敗しました',
        max_instances: 'インスタンス数の上限に達しました。しばらくしてから再度お試しください。',
      },
      action: {
        retry: '再試行',
        start_desktop: 'デスクトップを起動',
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
