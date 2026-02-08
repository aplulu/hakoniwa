import { Flex, Text } from '@radix-ui/themes';
import { Loader2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface LoadingScreenProps {
  title?: string;
}

export function LoadingScreen({ title }: LoadingScreenProps) {
  const { t } = useTranslation();

  return (
    <Flex align="center" justify="center" style={{ minHeight: '100vh' }}>
      <Flex direction="column" align="center" gap="4">
        <Loader2
          className="animate-spin"
          size={40}
          style={{ color: 'var(--gray-10)' }}
        />
        <Text color="gray">
          {t('status.connecting', { title: title || 'Hakoniwa' })}
        </Text>
      </Flex>
    </Flex>
  );
}
