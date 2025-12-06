import { useState } from 'react';
import { Flex, Box, Heading, Text, Grid, Card, AspectRatio, Button } from '@radix-ui/themes';
import { ArrowLeft, AlertCircle, Terminal, CheckCircle2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { InstanceType } from '../../types';

interface CreateInstanceViewProps {
  instanceTypes?: InstanceType[];
  isCreating: boolean;
  error?: string;
  onBack: () => void;
  onCreate: (typeId: string) => void;
}

export function CreateInstanceView({ instanceTypes, isCreating, error, onBack, onCreate }: CreateInstanceViewProps) {
  const { t } = useTranslation();
  const [selectedTypeId, setSelectedTypeId] = useState<string | null>(null);

  return (
    <Flex direction="column" gap="6" className="anim-entry" style={{ height: '100%' }}>
      {/* Header Section */}
      <Flex direction="column" gap="4">
        <Button variant="ghost" color="gray" onClick={onBack} style={{ alignSelf: 'start', padding: 0, height: 'auto', gap: '8px' }}>
          <ArrowLeft size={20} />
          {t('workspace.action.back_to_list')}
        </Button>
        <Box>
          <Heading size="8" weight="bold" style={{ letterSpacing: '-0.03em', marginBottom: '12px' }}>{t('workspace.create.modal_title')}</Heading>
          <Text size="3" color="gray">{t('workspace.create.modal_desc')}</Text>
        </Box>
      </Flex>

      {error && (
        <Box p="3" style={{ background: 'var(--red-3)', borderRadius: 'var(--radius-2)' }}>
          <Flex gap="2" align="center">
            <AlertCircle size={16} color="var(--red-11)" />
            <Text color="red" size="2" weight="medium">{error}</Text>
          </Flex>
        </Box>
      )}

      {/* Selection Grid */}
      <Grid columns={{ initial: '1', sm: '2', md: '3' }} gap="4">
        {instanceTypes?.map((type) => {
          const isSelected = selectedTypeId === type.id;
          return (
            <Box 
              key={type.id}
              onClick={() => setSelectedTypeId(type.id)}
              style={{ 
                cursor: isCreating ? 'wait' : 'pointer',
                position: 'relative',
                transition: 'all 0.2s ease'
              }}
            >
              <Card 
                size="3"
                style={{ 
                  padding: '16px',
                  border: isSelected ? '2px solid var(--accent-9)' : '2px solid transparent',
                  backgroundColor: isSelected ? 'var(--accent-2)' : 'var(--gray-2)',
                  height: '100%',
                }}
                className={isSelected ? 'selected-card' : 'hover-card'}
              >
                <Flex direction="column" gap="3">
                  <AspectRatio ratio={16/10}>
                    <Box 
                      style={{ 
                        width: '100%', 
                        height: '100%', 
                        background: 'var(--gray-4)',
                        borderRadius: 'var(--radius-3)',
                        overflow: 'hidden',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        position: 'relative'
                      }}
                    >
                      {type.logo_url ? (
                        <img 
                          src={type.logo_url} 
                          alt={type.name} 
                          style={{ width: '100%', height: '100%', objectFit: 'cover' }} 
                        />
                      ) : (
                        <Terminal size={48} color="var(--gray-8)" />
                      )}
                      
                      {isSelected && (
                        <Box 
                          style={{
                            position: 'absolute',
                            top: 8,
                            right: 8,
                            background: 'var(--accent-9)',
                            borderRadius: '50%',
                            padding: 4,
                            color: 'white',
                            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
                            lineHeight: 0,
                          }}
                        >
                          <CheckCircle2 size={16} />
                        </Box>
                      )}
                    </Box>
                  </AspectRatio>
                  
                  <Box>
                    <Heading size="4" mb="1" weight="medium">{type.name}</Heading>
                    {type.description && (
                      <Text size="2" color="gray" style={{ lineHeight: '1.5' }}>
                        {type.description}
                      </Text>
                    )}
                  </Box>
                </Flex>
              </Card>
            </Box>
          );
        })}
      </Grid>

      {/* Footer Action */}
      <Box 
        mt="auto" 
        pt="4" 
        style={{ 
          borderTop: '1px solid var(--gray-4)',
          display: 'flex',
          justifyContent: 'flex-end'
        }}
      >
        <Button 
          size="4" 
          disabled={!selectedTypeId || isCreating} 
          onClick={() => selectedTypeId && onCreate(selectedTypeId)}
          style={{ 
            width: '100%', 
            maxWidth: '240px', 
            fontWeight: 'bold' 
          }}
        >
          {isCreating ? t('workspace.status.creating') : t('workspace.create.submit')}
        </Button>
      </Box>
    </Flex>
  );
}
