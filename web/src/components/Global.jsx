import { createGlobalStyle } from 'styled-components';

const GlobalStyles = createGlobalStyle`
  a {
    color: inherit;
    text-decoration: underline;
    
    transition-duration: 0.2s;
    transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
    
    &:hover {
      color: #bf4e30;
    }
  }
`;

export default GlobalStyles;
