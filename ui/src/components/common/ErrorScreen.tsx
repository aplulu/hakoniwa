import { Flex, Card, Heading, Text, Box, Button } from '@radix-ui/themes';
import { AlertCircle } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface ErrorScreenProps {
  error: string;
  onRetry: () => void;
}

export function ErrorScreen({ error, onRetry }: ErrorScreenProps) {
  const { t } = useTranslation();

  return (
    <Flex
      align="center"
      justify="center"
      style={{ minHeight: '100vh' }}
      p="4">
      <Card size="3" style={{ width: '100%', maxWidth: 400 }}>
        <Flex gap="4" direction="column">
          <Flex gap="2" align="center">
            <AlertCircle color="var(--red-9)" />
            <Heading size="4" color="red">
              {t('error.title')}
            </Heading>
          </Flex>
          <Text size="2" color="gray">
            {t('error.generic_desc')}
          </Text>

          <Box
            p="3"
            style={{
              background: 'var(--gray-3)',
              borderRadius: 'var(--radius-2)',
            }}>
            <Text size="2" color="red">
              {error || t('error.connection_failed')}
            </Text>
          </Box>

          <Button
            color="red"
            variant="soft"
            onClick={onRetry}
            style={{ width: '100%' }}>
            {t('action.retry')}
          </Button>
        </Flex>
      </Card>
    </Flex>
  );
}
