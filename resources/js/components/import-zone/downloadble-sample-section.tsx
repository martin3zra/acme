import { Replacements } from '@/types';
import { Download } from 'lucide-react';
import { Button } from '../ui/button';

const PRODUCT_CSV_CONTENT = `Nombre,Descripción,Precio,ITBIS,TIPO,SKU,Codigo,Codigo de barra,Referencia,Referencia suplidor
"Laptop Pro 15","Computadora de alto rendimiento",1200.00,18,Producto,"LP-001","PRD-01","7410258963","REF-A1","SUP-X100"
"Servicio Técnico","Mantenimiento preventivo",50.00,18,Servicio,"ST-99","SRV-01","","REF-B2",""`;

const PRODUCT_TXT_CONTENT: string = `Nombre\tDescripción\tPrecio\tITBIS\tTIPO\tSKU\tCodigo\tCodigo de barra\tReferencia\tReferencia suplidor
"Laptop Pro 15"\t"Computadora de alto rendimiento"\t1200.00\t18\tProducto\t"LP-001"\t"PRD-01"\t"7410258963"\t"REF-A1"\t"SUP-X100"
"Servicio Técnico"\t"Mantenimiento preventivo"\t50.00\t18\tServicio\t"ST-99"\t"SRV-01"\t""\t"REF-B2"\t""`;

const CUSTOMER_TXT_CONTENT = `Nombre\tNombre del contacto\tTelefono\tCorreo Electronico\tMonto Vencido\tTIPO\tMetodo de Pago\tLimite de credito\tTermino de Pago\tTipo Comprobante Fiscal\tCodigo\tLimite de Credito (TRUE/FALSE)
"Ferretería Central"\t"Juan Pérez"\t"8095550123"\t"juan@central.com"\t1500.00\tNegocio\tck\t50000.00\t30\t"Crédito Fiscal"\t"CLI-001"\tTRUE
"María López"\t"María López"\t"8295559876"\t"maria.l@gmail.com"\t0.00\tIndividuo\tcash\t0.00\t0\t"Consumo"\t"CLI-002"\tFALSE`;

const CUSTOMER_CSV_CONTENT: string = `Nombre,Nombre del contacto,Telefono,Correo Electronico,Monto Vencido,TIPO,Metodo de Pago,Limite de credito,Termino de Pago,Tipo Comprobante Fiscal,Codigo,Limite de Credito (TRUE/FALSE)
"Ferretería Central","Juan Pérez","8095550123","juan@central.com",1500.00,Negocio,ck,50000.00,30,"Crédito Fiscal","CLI-001",TRUE
"María López","María López","8295559876","maria.l@gmail.com",0.00,Individuo,cash,0.00,0,"Consumo","CLI-002",FALSE`;

type Source = 'items' | 'customers';

const configs: Record<Source, { contentCsv: string; contentTxt: string; baseName: string }> = {
  items: {
    contentCsv: PRODUCT_CSV_CONTENT,
    contentTxt: PRODUCT_TXT_CONTENT,
    baseName: 'plantilla_productos',
  },
  customers: {
    contentCsv: CUSTOMER_CSV_CONTENT,
    contentTxt: CUSTOMER_TXT_CONTENT,
    baseName: 'plantilla_clientes',
  },
};

type Props = {
  source: Source;
  t: (key: string, replacements?: Replacements) => string;
};
export function DownloadableSampleSection({ t, source }: Props) {
  /**
   * Triggers a browser download for a given string content.
   * @param content - The string data (CSV, TXT, etc.)
   * @param filename - The name of the file including extension (e.g., 'data.csv')
   * @param type - The MIME type (e.g., 'text/csv' or 'text/plain')
   */
  const handleDownload = (content: string, filename: string, type: 'text/csv' | 'text/plain'): void => {
    const blob: Blob = new Blob([content], { type });
    const url: string = window.URL.createObjectURL(blob);

    const link: HTMLAnchorElement = document.createElement('a');
    link.href = url;
    link.download = filename;

    document.body.appendChild(link);
    link.click();

    // Cleanup
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  };

  return (
    <div className="mb-6 space-y-2">
      <p className="text-muted-foreground text-sm">{t('global.import.sampleFiles')}</p>

      <div className="flex space-x-3">
        <Button variant="outline" size="sm" onClick={() => handleDownload(configs[source].contentCsv, configs[source].baseName, 'text/csv')}>
          <Download className="mr-2 h-4 w-4" />
          {t('global.import.downloadCsv')}
        </Button>
        <Button variant="outline" size="sm" onClick={() => handleDownload(configs[source].contentCsv, configs[source].baseName, 'text/plain')}>
          <Download className="mr-2 h-4 w-4" />
          {t('global.import.downloadTxt')}
        </Button>
      </div>
    </div>
  );
}
