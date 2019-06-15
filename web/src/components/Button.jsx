import styled from 'styled-components';

export const Button = styled.button`
  min-width: 100px;
  /* height: 20px; */

  margin: 15px;
  padding: 5px 10px;

  background: none;
  border: 2px solid #4392f1;
  box-shadow: 0 2px 3px rgba(0, 0, 0, 0.2);
  color: #4392f1;
  cursor: pointer;
  letter-spacing: 2px;
  font-family: 'Oswald', sans-serif;
  font-size: 1.2em;

  transition-duration: 0.3s;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);

  &:hover,
  &.active {
    border-color: #bf4e30;
    box-shadow: 0 4px 5px rgba(0, 0, 0, 0.3);
    color: #bf4e30;
    transform: translate(0, -10%);
  }
`;
